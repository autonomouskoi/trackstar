package stagelinq

import (
	"context"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/icedream/go-stagelinq"
	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/web/webutil"
	"github.com/autonomouskoi/mapset"
	"github.com/autonomouskoi/trackstar"
)

const (
	appName    = "aktrackstarstagelinq"
	appVersion = "0.0.1"
	timeout    = time.Second * 5

	EnvLocalContentPath = "AK_TRACKSTAR_STAGELINQ_CONTENT"
)

//go:embed web.zip
var webZip []byte

var (
	cfgKVKey = []byte("config")
)

func init() {
	manifest := &modules.Manifest{
		Id:          "9567a0da6bf0061e",
		Name:        "trackstarstagelinq",
		Description: "Retrieve real-time track information from StagelinQ-capable Denon DJ devices",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "/m/trackstarstagelinq/",
				Description: "Stagelinq Controls",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_EMBED_CONTROL,
			},
		},
	}
	modules.Register(manifest, &StagelinQ{})
}

var stateValues = []string{
	stagelinq.EngineDeck1.ExternalMixerVolume(),
	stagelinq.EngineDeck1.Play(),
	stagelinq.EngineDeck1.PlayState(),
	stagelinq.EngineDeck1.PlayStatePath(),
	stagelinq.EngineDeck1.TrackArtistName(),
	stagelinq.EngineDeck1.TrackTrackNetworkPath(),
	stagelinq.EngineDeck1.TrackSongLoaded(),
	stagelinq.EngineDeck1.TrackSongName(),
	stagelinq.EngineDeck1.TrackTrackData(),
	stagelinq.EngineDeck1.TrackTrackName(),

	stagelinq.EngineDeck2.ExternalMixerVolume(),
	stagelinq.EngineDeck2.Play(),
	stagelinq.EngineDeck2.PlayState(),
	stagelinq.EngineDeck2.PlayStatePath(),
	stagelinq.EngineDeck2.TrackArtistName(),
	stagelinq.EngineDeck2.TrackTrackNetworkPath(),
	stagelinq.EngineDeck2.TrackSongLoaded(),
	stagelinq.EngineDeck2.TrackSongName(),
	stagelinq.EngineDeck2.TrackTrackData(),
	stagelinq.EngineDeck2.TrackTrackName(),

	stagelinq.EngineDeck3.ExternalMixerVolume(),
	stagelinq.EngineDeck3.Play(),
	stagelinq.EngineDeck3.PlayState(),
	stagelinq.EngineDeck3.PlayStatePath(),
	stagelinq.EngineDeck3.TrackArtistName(),
	stagelinq.EngineDeck3.TrackTrackNetworkPath(),
	stagelinq.EngineDeck3.TrackSongLoaded(),
	stagelinq.EngineDeck3.TrackSongName(),
	stagelinq.EngineDeck3.TrackTrackData(),
	stagelinq.EngineDeck3.TrackTrackName(),

	stagelinq.EngineDeck4.ExternalMixerVolume(),
	stagelinq.EngineDeck4.Play(),
	stagelinq.EngineDeck4.PlayState(),
	stagelinq.EngineDeck4.PlayStatePath(),
	stagelinq.EngineDeck4.TrackArtistName(),
	stagelinq.EngineDeck4.TrackTrackNetworkPath(),
	stagelinq.EngineDeck4.TrackSongLoaded(),
	stagelinq.EngineDeck4.TrackSongName(),
	stagelinq.EngineDeck4.TrackTrackData(),
	stagelinq.EngineDeck4.TrackTrackName(),
}

var (
	supportedDeviceNames = mapset.From("prime4", "sc5000", "sc6000")
	supportedSoftwares   = mapset.From("JC11", "JP07")
)

func makeStateMap() map[string]bool {
	retval := map[string]bool{}
	for _, value := range stateValues {
		retval[value] = false
	}
	return retval
}

func allStateValuesReceived(v map[string]bool) bool {
	for _, value := range v {
		if !value {
			return false
		}
	}
	return true
}

type deckState struct {
	deckID   string
	notified bool
	track    *trackstar.Track
	playing  bool
	upfader  float64
}

type StagelinQ struct {
	http.Handler
	bus        *bus.Bus
	cancel     func()
	log        *slog.Logger
	listener   *stagelinq.Listener
	discovered *discovered
	deckStates map[string]*deckState
	cfg        *Config
}

func (sl *StagelinQ) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	sl.log = deps.Log.With("module", "trackstar/stagelinq")
	sl.bus = deps.Bus
	sl.discovered = &discovered{}
	sl.deckStates = map[string]*deckState{}

	sl.cfg = &Config{}
	if cfgBytes, err := deps.KV.Get(cfgKVKey); err == nil {
		if err := proto.Unmarshal(cfgBytes, sl.cfg); err != nil {
			return fmt.Errorf("unmarshalling config: %w", err)
		}
	} else if !errors.Is(err, akcore.ErrNotFound) {
		return fmt.Errorf("retrieving config: %w", err)
	}

	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, webZip)
	if err != nil {
		return err
	}
	sl.Handler = http.StripPrefix("/m/trackstarstagelinq", http.FileServer(fs))

	defer func() {
		b, err := proto.Marshal(sl.cfg)
		if err != nil {
			sl.log.Error("marshalling config", "error", err.Error())
			return
		}
		if err := deps.KV.Set(cfgKVKey, b); err != nil {
			sl.log.Error("storing config", "error", err.Error())
			return
		}
	}()

	ctx, sl.cancel = context.WithCancel(ctx)
	defer sl.cancel()

	sl.listener, err = stagelinq.ListenWithConfiguration(&stagelinq.ListenerConfiguration{
		//Context:          ctx,
		DiscoveryTimeout: timeout,
		SoftwareName:     appName,
		SoftwareVersion:  appVersion,
		Name:             "boop",
	})
	if err != nil {
		return fmt.Errorf("listening for StagelinQ devices: %w", err)
	}
	defer sl.listener.Close()
	sl.listener.AnnounceEvery(time.Second)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			if err := ctx.Err(); err != nil {
				break
			}
			sl.discover(ctx)
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		sl.handleRequests(ctx)
		wg.Done()
	}()
	wg.Wait()
	return ctx.Err()
}

func (sl *StagelinQ) discover(ctx context.Context) {
	device, deviceState, err := sl.listener.Discover(timeout)
	if err != nil {
		sl.log.Error("discovering devices", "error", err.Error())
		return
	}
	if device == nil {
		return
	}
	if deviceState != stagelinq.DevicePresent {
		sl.log.Debug("discovered non-present device", "state", deviceState, "address", device.IP)
		return
	}
	if !supportedDeviceNames.Has(device.Name) {
		return
	}
	if !supportedSoftwares.Has(device.SoftwareName) {
		return
	}
	if sl.discovered.haveFound(device) {
		return
	}
	osl := *sl
	osl.log = sl.log.With(
		"ip", device.IP,
		"device_name", device.Name,
		"software_name", device.SoftwareName,
		"software_version", device.SoftwareVersion,
	)
	go func() {
		defer func() {
			if v := recover(); v != nil {
				stack := debug.Stack()
				osl.log.Error("panic", "v", v, "trace", string(stack))
				sl.cancel()
			}
		}()
		for ctx.Err() == nil {
			if err := osl.handleDevice(ctx, device); err != nil {
				osl.log.Error("handling device", "error", err.Error())
				time.Sleep(time.Second * 5) // wait 5 seconds before trying again
			}
		}
	}()
}

func (sl *StagelinQ) handleDevice(ctx context.Context, device *stagelinq.Device) error {
	token := sl.listener.Token()
	sl.log.Debug("handling device", "token", hex.EncodeToString(token[:]))
	deviceConn, err := device.Connect(token, []*stagelinq.Service{})
	if err != nil {
		return fmt.Errorf("connecting to device: %w", err)
	}
	defer deviceConn.Close()

	retryDelay := time.Millisecond * 50
	var services []*stagelinq.Service
	for i := 0; i < 6; i++ {
		sl.log.Debug("requesting data services", "ip", device.IP)
		services, err = deviceConn.RequestServices()
		if err != nil {
			return fmt.Errorf("requesting services: %w", err)
		}
		if len(services) > 0 {
			break
		}
		time.Sleep(retryDelay)
		retryDelay *= 2
	}
	if len(services) == 0 {
		return errors.New("no services discovered")
	}

	sl.discovered.setServices(device, services)
	sl.handleGetDevices(nil)

	for _, service := range services {
		sl.log.Debug("service offer",
			"name", service.Name,
			"port", service.Port,
		)
		if service.Name != "StateMap" {
			continue
		}
		smh := stateMapHandler{
			device:  device,
			service: service,
			token:   token,
		}
		for {
			if err := sl.handleStateMap(ctx, smh); err != nil {
				sl.log.Error("handling StateMap", "err", err.Error())
			}
		}
	}
	sl.log.Debug("finished device")

	return nil
}

type stateMapHandler struct {
	device  *stagelinq.Device
	service *stagelinq.Service
	token   stagelinq.Token
}

func (sl *StagelinQ) handleStateMap(ctx context.Context, smh stateMapHandler) error {
	sl.log.Debug("handling state map")
	stateMapTCPConn, err := smh.device.Dial(smh.service.Port)
	if err != nil {
		return fmt.Errorf("creating stateMapTCPConn: %w", err)
	}
	defer stateMapTCPConn.Close()
	stateMapConn, err := stagelinq.NewStateMapConnection(stateMapTCPConn, smh.token)
	if err != nil {
		return fmt.Errorf("creating stateMapConn: %w", err)
	}

	m := makeStateMap()
	for _, stateValue := range stateValues {
		stateMapConn.Subscribe(stateValue)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-stateMapConn.ErrorC():
			return fmt.Errorf("in state map connection: %w", err)
		case state := <-stateMapConn.StateC():
			sl.handleState(smh.device, state)
			m[state.Name] = true
			if allStateValuesReceived(m) {
				return nil
			}
		}
	}
}

func (sl *StagelinQ) handleState(device *stagelinq.Device, state *stagelinq.State) {
	nameFields := strings.Split(state.Name, "/")
	if len(nameFields) < 3 {
		return
	}
	deckID := device.IP.String() + "/" + nameFields[2]
	ds, present := sl.deckStates[deckID]
	if !present {
		sl.log.Debug("new deck", "deckID", deckID)
		b, err := proto.Marshal(&trackstar.DeckDiscovered{
			DeckId: deckID,
		})
		if err != nil {
			sl.log.Error("marshalling DeckDiscovered proto", "error", err.Error())
			return
		}
		sl.bus.Send(&bus.BusMessage{
			Topic:   trackstar.BusTopic_TRACKSTAR.String(),
			Type:    int32(trackstar.MessageType_TYPE_DECK_DISCOVERED),
			Message: b,
		})
		ds = &deckState{
			deckID: deckID,
			track:  &trackstar.Track{},
		}
		sl.deckStates[deckID] = ds
	}
	if nameFields[3] == "PlayState" {
		playing, ok := state.Value["state"].(bool)
		if !ok {
			return
		}
		ds.playing = playing
		sl.maybeNotify(ds)
		return
	}
	if nameFields[3] == "ExternalMixerVolume" {
		upfader, ok := state.Value["value"].(float64)
		if !ok {
			return
		}
		ds.upfader = upfader
		sl.maybeNotify(ds)
		return
	}
	if len(nameFields) < 5 {
		return
	}
	if nameFields[3] != "Track" {
		return
	}
	switch nameFields[4] {
	case "ArtistName":
		v, ok := state.Value["string"].(string)
		if !ok {
			return
		}
		ds.track.Artist = v
		ds.notified = false
	case "SongName":
		v, ok := state.Value["string"].(string)
		if !ok {
			return
		}
		ds.track.Title = v
		ds.notified = false
	default:
		return
	}
}

func (sl *StagelinQ) maybeNotify(ds *deckState) {
	if ds.notified {
		return
	}

	if !ds.playing {
		return
	}
	if ds.upfader <= sl.cfg.FaderThreshold && sl.cfg.FaderThreshold != 0 {
		return
	}

	b, err := proto.Marshal(&trackstar.TrackUpdate{
		DeckId: ds.deckID,
		Track:  ds.track,
		When:   time.Now().Unix(),
	})
	if err != nil {
		sl.log.Error("marshalling Track proto", "error", err.Error())
		return
	}
	sl.log.Debug("sending track", "track", ds.track)
	sl.bus.Send(&bus.BusMessage{
		Topic:   trackstar.BusTopic_TRACKSTAR.String(),
		Type:    int32(trackstar.MessageType_TYPE_TRACK_UPDATE),
		Message: b,
	})
	ds.notified = true
}

func (sl *StagelinQ) handleRequests(ctx context.Context) {
	in := make(chan *bus.BusMessage, 16)
	sl.bus.Subscribe(BusTopics_STAGELINQ_CONTROL.String(), in)
	defer func() {
		sl.bus.Unsubscribe(BusTopics_STAGELINQ_CONTROL.String(), in)
		for range in { // drain channel
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-in:
			switch msg.Type {
			case int32(MessageType_TYPE_CAPTURE_THRESHOLD_REQUEST):
				highest := 0.0
				for _, ds := range sl.deckStates {
					if ds.upfader > highest {
						highest = ds.upfader
					}
				}
				sl.cfg.FaderThreshold = highest
				sl.sendThreshold()
			case int32(MessageType_TYPE_GET_THRESHOLD_REQUEST):
				sl.sendThreshold()
			case int32(MessageType_TYPE_GET_DEVICES_REQUEST):
				sl.handleGetDevices(msg)
			}
		}
	}
}

func (sl *StagelinQ) sendThreshold() {
	b, err := proto.Marshal(&ThresholdUpdate{
		FaderThreshold: sl.cfg.FaderThreshold,
	})
	if err != nil {
		sl.log.Error("marshalling CaptureThresholdResponse", "error", err.Error())
		return
	}
	sl.bus.Send(&bus.BusMessage{
		Topic:   BusTopics_STAGELINQ_STATE.String(),
		Type:    int32(MessageType_TYPE_THRESHOLD_UPDATE),
		Message: b,
	})
}

func (sl *StagelinQ) handleGetDevices(msg *bus.BusMessage) {
	sl.discovered.processDevices(func(m []*device) {
		devices := make([]*Device, 0, len(m))
		for _, device := range m {
			devices = append(devices, device.pb)
		}
		slices.SortFunc(devices, func(a, b *Device) int {
			switch {
			case a.Ip == b.Ip:
				return 0
			case a.Ip == b.Ip:
				return -1
			}
			return 1
		})
		b, err := proto.Marshal(&GetDevicesResponse{
			Devices: devices,
		})
		if err != nil {
			sl.log.Error("marshalling GetDevicesResponse", "error", err.Error())
			return
		}
		reply := &bus.BusMessage{
			Topic:   BusTopics_STAGELINQ_STATE.String(),
			Type:    int32(MessageType_TYPE_GET_DEVICES_RESPONSE),
			Message: b,
		}
		if msg == nil {
			sl.bus.Send(reply)
		} else {
			sl.bus.SendReply(msg, reply)
		}
	})
}
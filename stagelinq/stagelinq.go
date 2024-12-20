package stagelinq

import (
	"context"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/icedream/go-stagelinq"
	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/storage/kv"
	"github.com/autonomouskoi/akcore/web/webutil"
	"github.com/autonomouskoi/datastruct/mapset"
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
				Path:        "https://autonomouskoi.org/module-trackstarstagelinq.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_HELP,
				Description: "Help!",
			},
			{
				Path:        "/m/trackstarstagelinq/embed_ctrl.js",
				Description: "Stagelinq Controls",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_EMBED_CONTROL,
			},
			{
				Path:        "/m/trackstarstagelinq/index.html",
				Description: "Stagelinq Controls",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_CONTROL_PAGE,
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
	modutil.ModuleBase
	bus        *bus.Bus
	cancel     func()
	listener   *stagelinq.Listener
	discovered *discovered
	deckStates map[string]*deckState
	cfg        *Config
	kv         kv.KVPrefix
}

func (sl *StagelinQ) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	sl.Log = deps.Log.With("module", "trackstar/stagelinq")
	sl.bus = deps.Bus
	sl.discovered = &discovered{}
	sl.deckStates = map[string]*deckState{}
	sl.kv = deps.KV

	sl.cfg = &Config{}

	if err := sl.kv.GetProto(cfgKVKey, sl.cfg); err != nil && !errors.Is(err, akcore.ErrNotFound) {
		return fmt.Errorf("retrieving config: %w", err)
	}
	defer sl.writeCfg()

	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, webZip)
	if err != nil {
		return err
	}
	sl.Handler = http.StripPrefix("/m/trackstarstagelinq", http.FileServer(fs))

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
	sl.Go(func() error {
		defer sl.Log.Debug("exiting", "loop", "discover")
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			sl.discover(ctx)
		}
	})
	sl.Go(func() error {
		sl.handleRequests(ctx)
		sl.Log.Debug("exiting", "loop", "handleRequests")
		return nil
	})
	sl.Wait()
	return ctx.Err()
}

func (sl *StagelinQ) writeCfg() {
	if err := sl.kv.SetProto(cfgKVKey, sl.cfg); err != nil {
		sl.Log.Error("writing config", "error", err.Error())
	}
}

func (sl *StagelinQ) discover(ctx context.Context) {
	device, deviceState, err := sl.listener.Discover(timeout)
	if err != nil {
		sl.Log.Error("discovering devices", "error", err.Error())
		return
	}
	if device == nil {
		return
	}
	if deviceState != stagelinq.DevicePresent {
		sl.Log.Debug("discovered non-present device", "state", deviceState, "address", device.IP)
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
	sl.Go(func() error {
		for ctx.Err() == nil {
			if err := sl.handleDevice(ctx, device); err != nil {
				sl.Log.Error("handling device", "error", err.Error())
				time.Sleep(time.Second * 5) // wait 5 seconds before trying again
			}
		}
		return nil
	})
}

func (sl *StagelinQ) handleDevice(ctx context.Context, device *stagelinq.Device) error {
	token := sl.listener.Token()
	sl.Log.Debug("handling device", "token", hex.EncodeToString(token[:]))
	deviceConn, err := device.Connect(token, []*stagelinq.Service{})
	if err != nil {
		return fmt.Errorf("connecting to device: %w", err)
	}
	defer deviceConn.Close()

	retryDelay := time.Millisecond * 50
	var services []*stagelinq.Service
	for i := 0; i < 6; i++ {
		sl.Log.Debug("requesting data services", "ip", device.IP)
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
		sl.Log.Debug("service offer",
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
		for ctx.Err() == nil {
			if err := sl.handleStateMap(ctx, smh); err != nil {
				sl.Log.Error("handling StateMap", "err", err.Error())
			}
		}
	}
	sl.Log.Debug("finished device")

	return nil
}

type stateMapHandler struct {
	device  *stagelinq.Device
	service *stagelinq.Service
	token   stagelinq.Token
}

func (sl *StagelinQ) handleStateMap(ctx context.Context, smh stateMapHandler) error {
	sl.Log.Debug("handling state map")
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
		sl.Log.Debug("new deck", "deckID", deckID)
		b, err := proto.Marshal(&trackstar.DeckDiscovered{
			DeckId: deckID,
		})
		if err != nil {
			sl.Log.Error("marshalling DeckDiscovered proto", "error", err.Error())
			return
		}
		sl.bus.Send(&bus.BusMessage{
			Topic:   trackstar.BusTopic_TRACKSTAR_EVENT.String(),
			Type:    int32(trackstar.MessageTypeEvent_TRACKSTAR_EVENT_DECK_DISCOVERED),
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
		sl.Log.Debug("track below threshold", "fader", ds.upfader, "threshold", sl.cfg.FaderThreshold)
		return
	}

	b, err := proto.Marshal(&trackstar.SubmitTrackRequest{
		TrackUpdate: &trackstar.TrackUpdate{
			DeckId: ds.deckID,
			Track:  ds.track,
			When:   time.Now().Unix(),
		}})
	if err != nil {
		sl.Log.Error("marshalling Track proto", "error", err.Error())
		return
	}
	sl.Log.Debug("sending track", "track", ds.track)
	sl.bus.Send(&bus.BusMessage{
		Topic:   trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
		Type:    int32(trackstar.MessageTypeRequest_SUBMIT_TRACK_REQ),
		Message: b,
	})
	ds.notified = true
}

func (sl *StagelinQ) handleRequests(ctx context.Context) {
	in := make(chan *bus.BusMessage, 16)
	sl.bus.Subscribe(BusTopics_STAGELINQ_CONTROL.String(), in)
	go func() {
		<-ctx.Done()
		sl.bus.Unsubscribe(BusTopics_STAGELINQ_CONTROL.String(), in)
		bus.Drain(in)
	}()
	for msg := range in {
		switch msg.Type {
		case int32(MessageType_TYPE_CAPTURE_THRESHOLD_REQUEST):
			highest := 0.0
			for _, ds := range sl.deckStates {
				if ds.upfader > highest {
					highest = ds.upfader
				}
			}
			sl.cfg.FaderThreshold = highest
			sl.Log.Debug("setting threshold", "threshold", highest)
			sl.writeCfg()
			sl.sendThreshold()
		case int32(MessageType_TYPE_GET_THRESHOLD_REQUEST):
			sl.sendThreshold()
		case int32(MessageType_TYPE_GET_DEVICES_REQUEST):
			sl.handleGetDevices(msg)
		}
	}
}

func (sl *StagelinQ) sendThreshold() {
	b, err := proto.Marshal(&ThresholdUpdate{
		FaderThreshold: sl.cfg.FaderThreshold,
	})
	if err != nil {
		sl.Log.Error("marshalling CaptureThresholdResponse", "error", err.Error())
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
			sl.Log.Error("marshalling GetDevicesResponse", "error", err.Error())
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

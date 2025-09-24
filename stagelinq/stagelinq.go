package stagelinq

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/icedream/go-stagelinq"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/storage/kv"
	"github.com/autonomouskoi/akcore/web/webutil"
	"github.com/autonomouskoi/datastruct/mapset"
	trackstar "github.com/autonomouskoi/trackstar/pb"
)

const (
	appName    = "aktrackstarstagelinq"
	appVersion = "0.0.13"
	timeout    = time.Second * 5

	EnvLocalContentPath = "AK_CONTENT_TRACKSTAR_STAGELINQ"
)

var (
	supportedDeviceNames = mapset.From("prime2", "prime4", "prime4+", "sc5000", "sc6000")
	supportedSoftwares   = mapset.From("JC11", "JP07")
)

//go:embed web.zip
var webZip []byte

var (
	cfgKVKey = []byte("config")
)

func init() {
	manifest := &modules.Manifest{
		Id:          "9567a0da6bf0061e",
		Title:       "TS: StagelinQ",
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
	bus          *bus.Bus
	listener     *stagelinq.Listener
	lock         sync.Mutex
	deviceStates *deviceStates
	deckStates   map[string]*deckState
	cfg          *Config
	kv           kv.KVPrefix
}

func (sl *StagelinQ) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	sl.Log = deps.Log
	sl.bus = deps.Bus
	sl.kv = deps.KV
	sl.deckStates = map[string]*deckState{}

	sl.cfg = &Config{}

	if err := sl.kv.GetProto(cfgKVKey, sl.cfg); err != nil && !errors.Is(err, akcore.ErrNotFound) {
		return fmt.Errorf("retrieving config: %w", err)
	}
	defer sl.writeCfg()

	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, webZip)
	if err != nil {
		return err
	}
	sl.Handler = http.FileServer(fs)

	sl.Go(func() error {
		sl.handleRequests(ctx)
		return nil
	})
	sl.Go(func() error {
		sl.handleCommands(ctx)
		return nil
	})
	sl.Go(func() error {
		defer sl.Log.Debug("exiting device search")
		for {
			sl.Log.Debug("starting device search")
			if err := sl.start(ctx); err != nil {
				sl.Log.Error("searching for devices", "error", err.Error())
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second * 15):
				// try again
			}
		}
	})
	return sl.Wait()
}

func (sl *StagelinQ) start(ctx context.Context) error {
	var err error
	sl.deviceStates = newDeviceStates(sl.bus, sl.Log)
	maps.Clear(sl.deckStates)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 15):
				if len(sl.deckStates) == 0 {
					cancel()
				}
			}
		}
	}()

	sl.listener, err = stagelinq.ListenWithConfiguration(&stagelinq.ListenerConfiguration{
		//Context:          ctx,
		DiscoveryTimeout: timeout,
		SoftwareName:     appName,
		SoftwareVersion:  appVersion,
		Name:             "AutonomousKoi",
	})
	if err != nil {
		return fmt.Errorf("listening for StagelinQ devices: %w", err)
	}
	defer sl.listener.Close()
	sl.listener.AnnounceEvery(time.Second)
	eg := errgroup.Group{}
	eg.Go(func() error {
		defer sl.Log.Debug("exiting", "loop", "discover")
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			sl.discover(ctx)
		}
	})
	eg.Go(func() error {
		sl.handleRequests(ctx)
		sl.Log.Debug("exiting", "loop", "handleRequests")
		return nil
	})
	return eg.Wait()
}

func (sl *StagelinQ) writeCfg() {
	sl.lock.Lock()
	defer sl.lock.Unlock()
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
	tsDevice, isNew := sl.deviceStates.discovered(device)
	if !isNew {
		return
	}
	if deviceState != stagelinq.DevicePresent {
		sl.Log.Debug("discovered non-present device", "state", deviceState, "address", device.IP)
		return
	}
	if !supportedDeviceNames.Has(device.Name) {
		sl.deviceStates.update(tsDevice.GetToken(),
			DeviceStatus_DEVICE_STATUS_UNSUPPORTED,
			"unsupported hardware",
		)
		sl.Log.Debug("unsupported device", "name", device.Name)
		return
	}
	if !supportedSoftwares.Has(device.SoftwareName) {
		sl.deviceStates.update(tsDevice.GetToken(),
			DeviceStatus_DEVICE_STATUS_UNSUPPORTED,
			"unsupported software",
		)
		sl.Log.Debug("unsupported software", "software_name", device.SoftwareName)
		return
	}
	sl.Go(func() error {
		if err := sl.handleDevice(ctx, device); err != nil {
			sl.Log.Error("handling device", "error", err.Error())
			sl.deviceStates.error(tsDevice.GetToken(), tsDevice.StatusDetail+": will retry")
		}
		return nil
	})
}

func (sl *StagelinQ) handleDevice(ctx context.Context, slDevice *stagelinq.Device) error {
	deviceToken := slDeviceToken(slDevice)
	sl.deviceStates.update(deviceToken, DeviceStatus_DEVICE_STATUS_CONNECTING, "Connecting...")
	deviceConn, err := slDevice.Connect(sl.listener.Token(), []*stagelinq.Service{})
	if err != nil {
		sl.deviceStates.error(deviceToken, "Connecting: "+err.Error())
		return fmt.Errorf("connecting to device: %w", err)
	}
	defer deviceConn.Close()
	sl.deviceStates.update(deviceToken, DeviceStatus_DEVICE_STATUS_INITIAL_CONNECT, "")

	retryDelay := time.Millisecond * 50
	var services []*stagelinq.Service
	for i := 0; i < 6; i++ {
		sl.Log.Debug("requesting data services", "ip", slDevice.IP, "token", deviceToken)
		services, err = deviceConn.RequestServices()
		if err != nil {
			sl.deviceStates.error(deviceToken, "Requesting services: "+err.Error())
			return fmt.Errorf("requesting services: %w", err)
		}
		if len(services) > 0 {
			break
		}
		time.Sleep(retryDelay)
		retryDelay *= 2
	}
	if len(services) == 0 {
		sl.deviceStates.error(deviceToken, "no services discovered")
		return errors.New("no services discovered")
	}

	sl.deviceStates.setServices(deviceToken, services)

	for _, service := range services {
		sl.Log.Debug("service offer",
			"name", service.Name,
			"port", service.Port,
		)
		if service.Name != "StateMap" {
			continue
		}
		smh := stateMapHandler{
			slDevice:      slDevice,
			deviceToken:   deviceToken,
			service:       service,
			listenerToken: sl.listener.Token(),
		}
		if err := sl.handleStateMap(ctx, smh); err != nil {
			sl.deviceStates.error(deviceToken, "Handling StateMap: "+err.Error())
			sl.Log.Error("handling StateMap", "err", err.Error())
		}
	}
	sl.Log.Debug("finished device")

	return nil
}

type stateMapHandler struct {
	slDevice      *stagelinq.Device
	service       *stagelinq.Service
	deviceToken   string
	listenerToken stagelinq.Token
}

func (sl *StagelinQ) handleStateMap(ctx context.Context, smh stateMapHandler) error {
	sl.Log.Debug("handling state map")
	defer maps.Clear(sl.deckStates)
	deviceToken := smh.deviceToken
	stateMapTCPConn, err := smh.slDevice.Dial(smh.service.Port)
	if err != nil {
		sl.deviceStates.error(deviceToken, "creating stateMapTCPConn: "+err.Error())
		return fmt.Errorf("creating stateMapTCPConn: %w", err)
	}
	defer stateMapTCPConn.Close()
	stateMapConn, err := stagelinq.NewStateMapConnection(stateMapTCPConn, smh.listenerToken)
	if err != nil {
		sl.deviceStates.error(deviceToken, "creating stateMapConn: "+err.Error())
		return fmt.Errorf("creating stateMapConn: %w", err)
	}

	m := makeStateMap()
	for _, stateValue := range stateValues {
		stateMapConn.Subscribe(stateValue)
	}
	sl.deviceStates.update(deviceToken, DeviceStatus_DEVICE_STATUS_READY, "Let's Rock!")
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-stateMapConn.ErrorC():
			return fmt.Errorf("in state map connection: %w", err)
		case state := <-stateMapConn.StateC():
			sl.handleState(smh.slDevice, state)
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
	//deckID := device.IP.String() + "/" + nameFields[2]
	deckID := nameFields[2]
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
			Type:    int32(trackstar.MessageTypeEvent_DECK_DISCOVERED),
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
	sl.Log.Info("sending track", "track", ds.track)
	sl.bus.Send(&bus.BusMessage{
		Topic:   trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
		Type:    int32(trackstar.MessageTypeRequest_SUBMIT_TRACK_REQ),
		Message: b,
	})
	ds.notified = true
}

//go:embed icon.svg
var icon []byte

func (*StagelinQ) Icon() ([]byte, string, error) {
	return icon, "image/svg+xml", nil
}

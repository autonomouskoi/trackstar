package trackstar

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/storage/kv"
	"github.com/autonomouskoi/akcore/web/webutil"
)

const (
	EnvLocalContentPath = "AK_TRACKSTAR_CONTENT"
)

var (
	cfgKVKey = []byte("config")
)

func init() {
	manifest := &modules.Manifest{
		Id:          "d6f95efeb3138d6e",
		Name:        "trackstar",
		Description: "Track songs played during your session",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "https://autonomouskoi.org/mod-trackstar.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_HELP,
				Description: "Help!",
			},
			{
				Path:        "/m/trackstar/",
				Type:        *modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_EMBED_CONTROL.Enum(),
				Description: "Configuration and track log",
			},
		},
	}
	modules.Register(manifest, &Trackstar{})
}

//go:embed web.zip
var webZip []byte

type Trackstar struct {
	http.Handler
	modutil.ModuleBase
	bus        *bus.Bus
	cfg        *Config
	kv         kv.KVPrefix
	lock       sync.Mutex
	updates    []*TrackUpdate
	demoCancel func()
}

func (ts *Trackstar) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	ts.bus = deps.Bus
	ts.Log = deps.Log
	ts.kv = deps.KV

	if err := ts.loadConfig(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	defer ts.writeCfg()

	ts.demoMode()

	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, webZip)
	if err != nil {
		return fmt.Errorf("get web FS %w", err)
	}
	ts.Handler = http.StripPrefix("/m/trackstar", http.FileServer(fs))

	eg := errgroup.Group{}
	eg.Go(func() error { return ts.handleRequests(ctx) })
	eg.Go(func() error { return ts.handleCommands(ctx) })

	return eg.Wait()
}

func (ts *Trackstar) handleRequests(ctx context.Context) error {
	ts.bus.HandleTypes(ctx, BusTopic_TRACKSTAR_REQUEST.String(), 8,
		map[int32]bus.MessageHandler{
			int32(MessageTypeRequest_TRACKSTAR_REQUEST_GET_TRACK_REQ): ts.handleGetTrackRequest,
			int32(MessageTypeRequest_SUBMIT_TRACK_REQ):                ts.handleRequestSubmitTrack,
			int32(MessageTypeRequest_CONFIG_GET_REQ):                  ts.handleRequestConfigGet,
			int32(MessageTypeRequest_GET_ALL_TRACKS_REQ):              ts.handleRequestGetAllTracks,
		},
		nil,
	)
	return nil
}

func (ts *Trackstar) handleGetTrackRequest(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	gtr := &GetTrackRequest{}
	if reply.Error = ts.UnmarshalMessage(msg, gtr); reply.Error != nil {
		return reply
	}
	ts.lock.Lock()
	defer ts.lock.Unlock()
	if len(ts.updates) == 0 {
		reply.Error = &bus.Error{
			Code:   int32(bus.CommonErrorCode_NOT_FOUND),
			Detail: proto.String("no tracks"),
		}
		return reply
	}
	when := time.Now().Add(-time.Second * time.Duration(gtr.DeltaSeconds)).Unix()
	gtResp := &GetTrackResponse{
		TrackUpdate: ts.updates[0],
	}
	for _, tu := range ts.updates[1:] {
		if tu.When > when {
			break
		}
		gtResp.TrackUpdate = tu
	}
	ts.MarshalMessage(reply, gtResp)
	return reply
}

var bracketRE = regexp.MustCompile(`\[.*\]`)
var multispaceRE = regexp.MustCompile(`\s{2,}`)

func (ts *Trackstar) mungeTrackUpdate(tu *TrackUpdate) {
	for match, replace := range ts.cfg.TrackReplacements {
		if strings.TrimSpace(match) == "" {
			continue
		}
		if strings.Contains(tu.Track.Artist, match) || strings.Contains(tu.Track.Title, match) {
			tu.Track.Artist = replace.Artist
			tu.Track.Title = replace.Title
			return
		}
	}
	if ts.cfg.ClearBracketedText {
		tu.Track.Artist = bracketRE.ReplaceAllString(tu.Track.Artist, " ")
		tu.Track.Artist = multispaceRE.ReplaceAllString(tu.Track.Artist, " ")
		tu.Track.Title = bracketRE.ReplaceAllString(tu.Track.Title, " ")
		tu.Track.Title = multispaceRE.ReplaceAllString(tu.Track.Title, " ")
	}
	tu.Track.Artist = strings.TrimSpace(tu.Track.Artist)
	tu.Track.Title = strings.TrimSpace(tu.Track.Title)
}

func (ts *Trackstar) handleRequestSubmitTrack(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.Topic,
		Type:  msg.Type + 1,
	}
	str := &SubmitTrackRequest{}
	if reply.Error = ts.UnmarshalMessage(msg, str); reply.Error != nil {
		return reply
	}
	ts.mungeTrackUpdate(str.TrackUpdate)

	go func() {
		time.Sleep(time.Second * time.Duration(ts.cfg.TrackDelaySeconds))
		ts.lock.Lock()
		ts.updates = append(ts.updates, str.TrackUpdate)
		ts.lock.Unlock()

		tuMsg := &bus.BusMessage{
			Topic: BusTopic_TRACKSTAR_EVENT.String(),
			Type:  int32(MessageTypeEvent_TRACKSTAR_EVENT_TRACK_UPDATE),
		}
		tuMsg.Message, _ = proto.Marshal(str.TrackUpdate)
		ts.Log.Debug("sending track", "deck_id", str.TrackUpdate.DeckId,
			"artist", str.TrackUpdate.Track.Artist,
			"title", str.TrackUpdate.Track.Title,
		)
		ts.bus.Send(tuMsg)
	}()

	ts.MarshalMessage(reply, &SubmitTrackResponse{})
	return reply
}

func (ts *Trackstar) handleRequestConfigGet(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	ts.lock.Lock()
	ts.MarshalMessage(reply, &ConfigGetResponse{
		Config: ts.cfg,
	})
	ts.lock.Unlock()
	return reply
}

func (ts *Trackstar) handleRequestGetAllTracks(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	ts.lock.Lock()
	ts.MarshalMessage(reply, &GetAllTracksResponse{
		Tracks: ts.updates,
	})
	ts.lock.Unlock()
	return reply
}

func (ts *Trackstar) handleCommands(ctx context.Context) error {
	ts.bus.HandleTypes(ctx, BusTopic_TRACKSTAR_COMMAND.String(), 4,
		map[int32]bus.MessageHandler{
			int32(MessageTypeCommand_CONFIG_SET_REQ): ts.handleCommandConfigSet,
		},
		nil,
	)
	return nil
}

func (ts *Trackstar) handleCommandConfigSet(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	csr := &ConfigSetRequest{}
	if reply.Error = ts.UnmarshalMessage(msg, csr); reply.Error != nil {
		return reply
	}
	ts.lock.Lock()
	demoDiffers := csr.GetConfig().GetDemoDelaySeconds() != ts.cfg.GetDemoDelaySeconds()
	ts.cfg = csr.GetConfig()
	ts.lock.Unlock()
	ts.writeCfg()
	ts.MarshalMessage(reply, &ConfigSetResponse{
		Config: ts.cfg,
	})
	if demoDiffers {
		if ts.demoCancel != nil {
			ts.demoCancel()
		}
		ts.demoMode()
	}
	return reply
}

func (ts *Trackstar) loadConfig() error {
	ts.cfg = &Config{}
	if err := ts.kv.GetProto(cfgKVKey, ts.cfg); err != nil && !errors.Is(err, akcore.ErrNotFound) {
		return fmt.Errorf("retrieving config: %w", err)
	}
	return nil
}

func (ts *Trackstar) writeCfg() {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	if err := ts.kv.SetProto(cfgKVKey, ts.cfg); err != nil {
		ts.Log.Error("writing config", "error", err.Error())
	}
}

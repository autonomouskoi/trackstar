package twitchchat

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/storage/kv"
	"github.com/autonomouskoi/akcore/web/webutil"
	"github.com/autonomouskoi/trackstar"
	"github.com/autonomouskoi/twitch"
)

const (
	EnvLocalContentPath = "AK_CONTENT_TRACKSTAR_TWITCHCHAT"

	defaultTemplate = `{{ .TrackUpdate.Track.Artist }} - {{ .TrackUpdate.Track.Title }}`
)

var (
	cfgKVKey = []byte("config")
)

func init() {
	manifest := &modules.Manifest{
		Id:          "62071945ac98ada1",
		Title:       "TS: Twitch Chat",
		Name:        "trackstartwitchchat",
		Description: "Trackstar integration with Twitch chat",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "https://autonomouskoi.org/module-trackstartwitchchat.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_HELP,
				Description: "Help!",
			},
			{
				Path:        "/m/trackstartwitchchat/embed_ctrl.js",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_EMBED_CONTROL,
				Description: "Configuration",
			},
			{
				Path:        "/m/trackstartwitchchat/index.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_CONTROL_PAGE,
				Description: "Configuration",
			},
		},
	}
	modules.Register(manifest, &TwitchChat{})
}

//go:embed web.zip
var webZip []byte

type TwitchChat struct {
	http.Handler
	modutil.ModuleBase
	bus  *bus.Bus
	lock sync.Mutex
	Log  *slog.Logger
	cfg  *Config
	kv   *kv.KVPrefix
	tmpl *template.Template
}

func (tc *TwitchChat) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, webZip)
	if err != nil {
		return err
	}
	tc.Handler = http.FileServer(fs)
	tc.bus = deps.Bus
	tc.kv = &deps.KV
	tc.Log = deps.Log

	tc.cfg = &Config{}
	if err := tc.kv.GetProto(cfgKVKey, tc.cfg); err != nil && !errors.Is(err, akcore.ErrNotFound) {
		return fmt.Errorf("retrieving config: %w", err)
	}
	if tc.cfg.Template == "" {
		tc.cfg.Template = defaultTemplate
	}
	if tc.tmpl, err = template.New("").Parse(tc.cfg.GetTemplate()); err != nil {
		tc.Log.Error("parsing template", "error", err.Error())
		tc.tmpl, _ = template.New("").Parse(defaultTemplate)
	}
	defer tc.writeCfg()

	eg := errgroup.Group{}
	eg.Go(func() error { return tc.handleChatMessagesIn(ctx) })
	eg.Go(func() error { return tc.handleCommands(ctx) })
	eg.Go(func() error { return tc.handleRequests(ctx) })
	eg.Go(func() error { return tc.handleTracks(ctx) })

	return eg.Wait()
}

func (tc *TwitchChat) handleCommands(ctx context.Context) error {
	tc.bus.HandleTypes(ctx, BusTopics_TRACKSTAR_TWITCH_CHAT_COMMAND.String(), 4,
		map[int32]bus.MessageHandler{
			int32(MessageTypeCommand_TRAKCSTAR_TWITCH_CHAT_CONFIG_SET_REQ): tc.handleCommandConfigSet,
		},
		nil)
	return nil
}

func (tc *TwitchChat) handleCommandConfigSet(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: BusTopics_TRACKSTAR_TWITCH_CHAT_COMMAND.String(),
		Type:  int32(MessageTypeCommand_TRAKCSTAR_TWITCH_CHAT_CONFIG_SET_RESP),
	}
	csr := &ConfigSetRequest{}
	if reply.Error = tc.UnmarshalMessage(msg, csr); reply.Error != nil {
		return reply
	}

	tmpl, err := template.New("").Parse(csr.Config.GetTemplate())
	if err != nil {
		reply.Error = &bus.Error{
			Code:   int32(bus.CommonErrorCode_INVALID_TYPE),
			Detail: proto.String("parsing template: " + err.Error()),
		}
		tc.Log.Error("parsing template", "error", err.Error())
		return reply
	}

	tc.lock.Lock()
	tc.cfg = csr.Config
	tc.tmpl = tmpl
	tc.lock.Unlock()
	tc.writeCfg()

	tc.MarshalMessage(reply, &ConfigSetResponse{})
	return reply
}

func (tc *TwitchChat) handleRequests(ctx context.Context) error {
	tc.bus.HandleTypes(ctx, BusTopics_TRACKSTAR_TWITCH_CHAT_REQUEST.String(), 4,
		map[int32]bus.MessageHandler{
			int32(MessageTypeRequest_TRACKSTAR_TWITCH_CHAT_CONFIG_GET_REQ):     tc.handleRequestConfigGet,
			int32(MessageTypeRequest_TRACKSTAR_TWITCH_CHAT_TRACK_ANNOUNCE_REQ): tc.handleRequestTrackAnnounce,
		},
		nil)
	return nil
}

func (tc *TwitchChat) handleRequestConfigGet(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.Topic,
		Type:  msg.Type + 1,
	}
	tc.lock.Lock()
	tc.MarshalMessage(reply, &ConfigGetResponse{
		Config: tc.cfg,
	})
	tc.lock.Unlock()
	return reply
}

func (tc *TwitchChat) handleRequestTrackAnnounce(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.Topic,
		Type:  msg.Type + 1,
	}
	tsReq := &bus.BusMessage{
		Topic: trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
		Type:  int32(trackstar.MessageTypeRequest_TRACKSTAR_REQUEST_GET_TRACK_REQ),
	}
	if tc.MarshalMessage(tsReq, &trackstar.GetTrackRequest{}); tsReq.Error != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	tsReply := tc.bus.WaitForReply(ctx, tsReq)
	if tsReply.Error != nil {
		reply.Error = tsReply.Error
		return reply
	}
	gtr := &trackstar.GetTrackResponse{}
	if reply.Error = tc.UnmarshalMessage(tsReply, gtr); reply.Error != nil {
		return reply
	}
	tc.sendTrackUpdate(gtr.GetTrackUpdate())
	return reply
}

func (tc *TwitchChat) handleChatMessagesIn(ctx context.Context) error {
	tc.bus.HandleTypes(ctx, twitch.BusTopics_TWITCH_EVENTSUB_EVENT.String(), 8,
		map[int32]bus.MessageHandler{
			int32(twitch.MessageTypeEventSub_TYPE_CHANNEL_CHAT_MESSAGE): tc.handleChatMessage,
		},
		nil)
	return nil
}

func (tc *TwitchChat) handleChatMessage(msg *bus.BusMessage) *bus.BusMessage {
	ccm := &twitch.EventChannelChatMessage{}
	if err := tc.UnmarshalMessage(msg, ccm); err != nil {
		return nil
	}
	text := strings.ToLower(ccm.GetMessage().Text)
	if !strings.HasPrefix(text, "!id") {
		return nil
	}

	req := &trackstar.GetTrackRequest{}
	for _, arg := range strings.Split(text, " ")[1:] {
		arg = strings.TrimSpace(arg)
		if arg != "" {
			text = arg
			break
		}
	}
	if text != "" {
		duration, err := time.ParseDuration(text)
		if err == nil {
			req.DeltaSeconds = uint32(duration / time.Second)
		}
	}

	reqMsg := &bus.BusMessage{
		Topic: trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
		Type:  int32(trackstar.MessageTypeRequest_TRACKSTAR_REQUEST_GET_TRACK_REQ),
	}
	if tc.MarshalMessage(reqMsg, req); reqMsg.Error != nil {
		return nil
	}
	reqCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	reply := tc.bus.WaitForReply(reqCtx, reqMsg)
	cancel()
	if reply.Error != nil {
		return nil
	}
	gtr := &trackstar.GetTrackResponse{}
	if err := tc.UnmarshalMessage(reply, gtr); err != nil {
		return nil
	}
	tc.sendTrackUpdate(gtr.TrackUpdate)
	return nil
}

func (tc *TwitchChat) sendTrackUpdate(tu *trackstar.TrackUpdate) {
	msg := &bus.BusMessage{
		Topic: twitch.BusTopics_TWITCH_CHAT_REQUEST.String(),
		Type:  int32(twitch.MessageTypeTwitchChatRequest_TWITCH_CHAT_REQUEST_TYPE_SEND_REQ),
	}
	buf := &bytes.Buffer{}
	err := tc.tmpl.Execute(buf, map[string]any{
		"TrackUpdate": tu,
	})
	if err != nil {
		tc.Log.Error("executing template", "error", err.Error())
		return
	}
	tc.MarshalMessage(msg, &twitch.TwitchChatRequestSendRequest{
		Text: buf.String(),
	})
	if msg.Error != nil {
		return
	}
	tc.bus.Send(msg)
}

func (tc *TwitchChat) handleTracks(ctx context.Context) error {
	tc.bus.HandleTypes(ctx, trackstar.BusTopic_TRACKSTAR_EVENT.String(), 8,
		map[int32]bus.MessageHandler{
			int32(trackstar.MessageTypeEvent_TRACKSTAR_EVENT_TRACK_UPDATE): tc.handleTrackUpdate,
		},
		nil)
	return nil
}

func (tc *TwitchChat) handleTrackUpdate(msg *bus.BusMessage) *bus.BusMessage {
	if !tc.cfg.Announce {
		return nil
	}
	tu := &trackstar.TrackUpdate{}
	if err := tc.UnmarshalMessage(msg, tu); err != nil {
		return nil
	}
	tc.sendTrackUpdate(tu)
	return nil
}

func (tc *TwitchChat) writeCfg() {
	tc.lock.Lock()
	defer tc.lock.Unlock()
	if err := tc.kv.SetProto(cfgKVKey, tc.cfg); err != nil {
		tc.Log.Error("writing config", "error", err.Error())
	}
}

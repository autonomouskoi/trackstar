package twitchchat

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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
	"github.com/autonomouskoi/trackstar"
	"github.com/autonomouskoi/twitch"
)

const (
	EnvLocalContentPath = "AK_TRACKSTAR_TWITCHCHAT_CONTENT"
)

var (
	cfgKVKey = []byte("config")
)

func init() {
	manifest := &modules.Manifest{
		Id:          "62071945ac98ada1",
		Name:        "trackstartwitchchat",
		Description: "Trackstar integration with Twitch chat",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "/m/trackstartwitchchat/",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_EMBED_CONTROL,
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
	bus  *bus.Bus
	lock sync.Mutex
	log  *slog.Logger
	cfg  *Config
	kv   *kv.KVPrefix
}

func (tc *TwitchChat) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, webZip)
	if err != nil {
		return err
	}
	tc.Handler = http.StripPrefix("/m/trackstartwitchchat", http.FileServer(fs))
	tc.bus = deps.Bus
	tc.kv = &deps.KV
	tc.log = deps.Log

	tc.cfg = &Config{}
	if err := tc.kv.GetProto(cfgKVKey, tc.cfg); err != nil && !errors.Is(err, akcore.ErrNotFound) {
		return fmt.Errorf("retrieving config: %w", err)
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
	in := make(chan *bus.BusMessage, 8)
	tc.bus.Subscribe(BusTopics_TRACKSTAR_TWITCH_CHAT_COMMAND.String(), in)
	go func() {
		<-ctx.Done()
		tc.bus.Unsubscribe(BusTopics_TRACKSTAR_TWITCH_CHAT_COMMAND.String(), in)
		bus.Drain(in)
	}()
	for msg := range in {
		var reply *bus.BusMessage
		switch msg.Type {
		case int32(MessageTypeCommand_TRAKCSTAR_TWITCH_CHAT_CONFIG_SET_REQ):
			reply = tc.handleCommandConfigSet(msg)
		}
		if reply != nil {
			tc.bus.SendReply(msg, reply)
		}
	}
	return nil
}

func (tc *TwitchChat) handleCommandConfigSet(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: BusTopics_TRACKSTAR_TWITCH_CHAT_COMMAND.String(),
		Type:  int32(MessageTypeCommand_TRAKCSTAR_TWITCH_CHAT_CONFIG_SET_RESP),
	}
	csr := &ConfigSetRequest{}
	if err := proto.Unmarshal(msg.GetMessage(), csr); err != nil {
		reply.Error = &bus.Error{
			Code:   int32(bus.CommonErrorCode_INVALID_TYPE),
			Detail: proto.String("unmarshalling " + err.Error()),
		}
		return reply
	}
	tc.lock.Lock()
	tc.cfg = csr.Config
	tc.lock.Unlock()
	tc.writeCfg()

	reply.Message, _ = proto.Marshal(&ConfigSetResponse{})
	return reply
}

func (tc *TwitchChat) handleRequests(ctx context.Context) error {
	in := make(chan *bus.BusMessage, 8)
	tc.bus.Subscribe(BusTopics_TRACKSTAR_TWITCH_CHAT_REQUEST.String(), in)
	go func() {
		<-ctx.Done()
		tc.bus.Unsubscribe(BusTopics_TRACKSTAR_TWITCH_CHAT_REQUEST.String(), in)
		bus.Drain(in)
	}()
	for msg := range in {
		var reply *bus.BusMessage
		switch msg.Type {
		case int32(MessageTypeRequest_TRACKSTAR_TWITCH_CHAT_CONFIG_GET_REQ):
			reply = tc.handleRequestConfigGet(msg)
		case int32(MessageTypeRequest_TRACKSTAR_TWITCH_CHAT_TRACK_ANNOUNCE_REQ):
			reply = tc.handleRequestTrackAnnounce(msg)
		}
		if reply != nil {
			tc.bus.SendReply(msg, reply)
		}
	}
	return nil
}

func (tc *TwitchChat) handleRequestConfigGet(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.Topic,
		Type:  msg.Type + 1,
	}
	tc.lock.Lock()
	reply.Message, _ = proto.Marshal(&ConfigGetResponse{
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
	tsReq.Message, _ = proto.Marshal(&trackstar.GetTrackRequest{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	tsReply := tc.bus.WaitForReply(ctx, tsReq)
	if tsReply.Error != nil {
		reply.Error = tsReply.Error
		return reply
	}
	gtr := &trackstar.GetTrackResponse{}
	if err := proto.Unmarshal(tsReply.GetMessage(), gtr); err != nil {
		tc.log.Error("unmarshalling", "type", "GetTrackResponse", "error", err.Error())
		reply.Error = &bus.Error{
			Code:   int32(bus.CommonErrorCode_INVALID_TYPE),
			Detail: proto.String(err.Error()),
		}
		return reply
	}
	tc.sendTrackUpdate(gtr.GetTrackUpdate())
	return reply
}

func (tc *TwitchChat) handleChatMessagesIn(ctx context.Context) error {
	in := make(chan *bus.BusMessage, 16)
	tc.bus.Subscribe(twitch.BusTopics_TWITCH_CHAT_RECV.String(), in)
	go func() {
		<-ctx.Done()
		tc.bus.Unsubscribe(twitch.BusTopics_TWITCH_CHAT_RECV.String(), in)
		bus.Drain(in)
	}()
	for msg := range in {
		cmi := &twitch.ChatMessageIn{}
		if err := proto.Unmarshal(msg.GetMessage(), cmi); err != nil {
			tc.log.Error("unmarshalling", "type", "ChatMessageIn", "error", err.Error())
			continue
		}
		text := strings.ToLower(cmi.Text)
		if !strings.HasPrefix(text, "!id") {
			continue
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

		b, err := proto.Marshal(req)
		if err != nil {
			tc.log.Error("marshalling", "type", "GetTrackRequest", "error", err.Error())
			continue
		}
		reqCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		reply := tc.bus.WaitForReply(reqCtx, &bus.BusMessage{
			Topic:   trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
			Type:    int32(trackstar.MessageTypeRequest_TRACKSTAR_REQUEST_GET_TRACK_REQ),
			Message: b,
		})
		cancel()
		if reply.Error != nil {
			tc.log.Error("requesting track", "code", reply.Error.GetCode(), "detail", reply.Error.GetDetail())
			continue
		}
		gtr := &trackstar.GetTrackResponse{}
		if err := proto.Unmarshal(reply.Message, gtr); err != nil {
			tc.log.Error("unmarshalling", "type", "GetTrackResponse", "error", err.Error())
			continue
		}
		tc.sendTrackUpdate(gtr.TrackUpdate)
	}
	return nil
}

func (tc *TwitchChat) sendTrackUpdate(tu *trackstar.TrackUpdate) {
	b, err := proto.Marshal(&twitch.ChatMessageOut{
		Text: fmt.Sprintf("%s - %s", tu.Track.Artist, tu.Track.Title),
	})
	if err != nil {
		tc.log.Error("marshalling", "type", "ChatMessageOut", "error", err.Error())
		return
	}
	tc.bus.Send(&bus.BusMessage{
		Topic:   twitch.BusTopics_TWITCH_CHAT_SEND.String(),
		Message: b,
	})
}

func (tc *TwitchChat) handleTracks(ctx context.Context) error {
	in := make(chan *bus.BusMessage, 8)
	tc.bus.Subscribe(trackstar.BusTopic_TRACKSTAR_EVENT.String(), in)
	go func() {
		<-ctx.Done()
		tc.bus.Unsubscribe(trackstar.BusTopic_TRACKSTAR_EVENT.String(), in)
		bus.Drain(in)
	}()
	for msg := range in {
		if msg.Type != int32(trackstar.MessageTypeEvent_TRACKSTAR_EVENT_TRACK_UPDATE) {
			continue
		}
		if !tc.cfg.Announce {
			continue
		}
		tu := &trackstar.TrackUpdate{}
		if err := proto.Unmarshal(msg.GetMessage(), tu); err != nil {
			tc.log.Error("unmarshalling", "type", "TrackUpdate", "error", err.Error())
			continue
		}
		tc.sendTrackUpdate(tu)
	}
	return nil
}

func (tc *TwitchChat) writeCfg() {
	tc.lock.Lock()
	defer tc.lock.Unlock()
	if err := tc.kv.SetProto(cfgKVKey, tc.cfg); err != nil {
		tc.log.Error("writing config", "error", err.Error())
	}
}

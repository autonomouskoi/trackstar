package trackstar

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/web/webutil"
)

const (
	EnvLocalContentPath = "AK_TRACKSTAR_CONTENT"
)

func init() {
	manifest := &modules.Manifest{
		Id:          "d6f95efeb3138d6e",
		Name:        "trackstar",
		Description: "Track songs played during your session",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "/m/trackstar/",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_GENERAL,
				Description: "Log of discovered tracks",
			},
		},
	}
	modules.Register(manifest, &Trackstar{})
}

//go:embed web.zip
var webZip []byte

type Trackstar struct {
	http.Handler
	bus     *bus.Bus
	lock    sync.Mutex
	log     *slog.Logger
	updates []*TrackUpdate
}

func (ts *Trackstar) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	ts.bus = deps.Bus
	ts.log = deps.Log

	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, webZip)
	if err != nil {
		return fmt.Errorf("get web FS %w", err)
	}
	ts.Handler = http.StripPrefix("/m/trackstar", http.FileServer(fs))

	eg := errgroup.Group{}
	eg.Go(func() error { return ts.handleEvents(ctx) })
	eg.Go(func() error { return ts.handleRequests(ctx) })

	return eg.Wait()
}

func (ts *Trackstar) handleEvents(ctx context.Context) error {
	in := make(chan *bus.BusMessage, 16)
	go func() {
		<-ctx.Done()
		ts.bus.Unsubscribe(BusTopic_TRACKSTAR_EVENT.String(), in)
		bus.Drain(in)
	}()
	ts.bus.Subscribe(BusTopic_TRACKSTAR_EVENT.String(), in)
	for msg := range in {
		switch msg.Type {
		case int32(MessageTypeEvent_TRACKSTAR_EVENT_TRACK_UPDATE):
			tu := &TrackUpdate{}
			if err := proto.Unmarshal(msg.Message, tu); err != nil {
				ts.log.Error("unmarshalling TrackUpdate", "error", err.Error())
				continue
			}
			ts.lock.Lock()
			ts.updates = append(ts.updates, tu)
			ts.lock.Unlock()
		}
	}
	return nil
}

func (ts *Trackstar) handleRequests(ctx context.Context) error {
	in := make(chan *bus.BusMessage, 8)
	ts.bus.Subscribe(BusTopic_TRACKSTAR_REQUEST.String(), in)
	go func() {
		<-ctx.Done()
		ts.bus.Unsubscribe(BusTopic_TRACKSTAR_REQUEST.String(), in)
		bus.Drain(in)
	}()
	for msg := range in {
		var reply *bus.BusMessage
		switch msg.Type {
		case int32(MessageTypeRequest_TRACKSTAR_REQUEST_GET_TRACK_REQ):
			reply = ts.handleGetTrackRequest(msg)
		}
		if reply != nil {
			ts.bus.SendReply(msg, reply)
		}
	}
	return nil
}

func (ts *Trackstar) handleGetTrackRequest(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  int32(MessageTypeRequest_TRACKSTAR_REQUEST_GET_TRACK_RESP),
	}
	gtr := &GetTrackRequest{}
	if err := proto.Unmarshal(msg.Message, gtr); err != nil {
		ts.log.Error("unmarshalling", "type", "GetTrackRequest", "error", err.Error())
		reply.Error = &bus.Error{
			Code:   int32(bus.CommonErrorCode_INVALID_TYPE),
			Detail: proto.String("unmarshalling GetTrackRequest: " + err.Error()),
		}
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
	b, err := proto.Marshal(gtResp)
	if err != nil {
		ts.log.Error("marshalling", "type", "GetTrackResponse", "error", err.Error())
		reply.Error = &bus.Error{
			Code:   int32(bus.CommonErrorCode_INVALID_TYPE),
			Detail: proto.String("marshalling GetTrackResponse: " + err.Error()),
		}
	}
	reply.Message = b
	return reply
}

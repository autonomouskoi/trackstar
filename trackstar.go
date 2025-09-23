package trackstar

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/storage/kv"
	"github.com/autonomouskoi/akcore/web/webutil"
	"github.com/autonomouskoi/trackstar/pb"
)

const (
	EnvLocalContentPath = "AK_CONTENT_TRACKSTAR"

	moduleID = "d6f95efeb3138d6e"
)

var (
	cfgKVKey = []byte("config")

	sessionPrefix = "session/"
)

func init() {
	manifest := &modules.Manifest{
		Id:          moduleID,
		Title:       "Trackstar",
		Name:        "trackstar",
		Description: "Track songs played during your session",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "https://autonomouskoi.org/module-trackstar.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_HELP,
				Description: "Help!",
			},
			{
				Path:        "/m/trackstar/embed_ctrl.js",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_EMBED_CONTROL,
				Description: "Configuration and track log",
			},
			{
				Path:        "/m/trackstar/index.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_CONTROL_PAGE,
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
	cfg        *pb.Config
	kv         kv.KVPrefix
	lock       sync.Mutex
	session    *pb.Session
	demoCancel func()
}

func (ts *Trackstar) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	ts.bus = deps.Bus
	ts.Log = deps.Log
	ts.kv = deps.KV
	ts.session = &pb.Session{Started: time.Now().Unix()}

	if err := ts.loadConfig(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	defer ts.writeCfg()

	ts.demoMode()

	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, webZip)
	if err != nil {
		return fmt.Errorf("get web FS %w", err)
	}
	ts.Handler = http.FileServer(fs)

	ts.Go(func() error { return ts.handleRequests(ctx) })
	ts.Go(func() error { return ts.handleCommands(ctx) })
	ts.Go(func() error { return ts.handleDirect(ctx) })

	return ts.Wait()
}

func (ts *Trackstar) handleCommands(ctx context.Context) error {
	ts.bus.HandleTypes(ctx, pb.BusTopic_TRACKSTAR_COMMAND.String(), 4,
		map[int32]bus.MessageHandler{
			int32(pb.MessageTypeCommand_CONFIG_SET_REQ): ts.handleCommandConfigSet,
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
	csr := &pb.ConfigSetRequest{}
	if reply.Error = ts.UnmarshalMessage(msg, csr); reply.Error != nil {
		return reply
	}
	if err := csr.Config.Validate(); err != nil {
		reply.Error = &bus.Error{
			Code:   int32(bus.CommonErrorCode_INVALID_TYPE),
			Detail: proto.String(err.Error()),
		}
		return reply
	}

	ts.lock.Lock()
	demoDiffers := csr.GetConfig().GetDemoDelaySeconds() != ts.cfg.GetDemoDelaySeconds()

	ts.cfg = csr.GetConfig()
	ts.lock.Unlock()
	ts.writeCfg()
	ts.MarshalMessage(reply, &pb.ConfigSetResponse{
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
	ts.cfg = &pb.Config{}
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

func (ts *Trackstar) saveSession(session *pb.Session) error {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	sessionKey := fmt.Sprintf("%s%d", sessionPrefix, session.GetStarted())
	return ts.kv.SetProto([]byte(sessionKey), session)
}

//go:embed icon.svg
var icon []byte

func (*Trackstar) Icon() ([]byte, string, error) {
	return icon, "image/svg+xml", nil
}

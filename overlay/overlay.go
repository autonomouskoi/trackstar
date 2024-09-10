package overlay

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/storage/kv"
	"github.com/autonomouskoi/akcore/web/webutil"
)

const (
	EnvLocalContentPath = "AK_TRACKSTAR_OVERLAY_CONTENT"
)

var (
	cfgKVKey = []byte("config")
)

func init() {
	manifest := &modules.Manifest{
		Id:          "74623a194d49d3ca",
		Name:        "trackstaroverlay",
		Description: "OBS Overlay for Trackstar",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "/m/trackstaroverlay/",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_OBS_OVERLAY,
				Description: "OBS Overlay",
			},
			{
				Path:        "/m/trackstaroverlay/ui.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_EMBED_CONTROL,
				Description: "Overlay Configuration",
			},
		},
	}
	modules.Register(manifest, &Overlay{})
}

type Overlay struct {
	http.Handler
	deps *modutil.ModuleDeps
	lock sync.Mutex
	log  *slog.Logger
	cfg  *Config
	kv   *kv.KVPrefix
}

func (o *Overlay) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, overlayHTML)
	if err != nil {
		return err
	}
	o.Handler = http.StripPrefix("/m/trackstaroverlay", http.FileServer(fs))
	o.deps = deps
	o.kv = &deps.KV
	o.log = deps.Log

	o.cfg = &Config{}
	if err := o.kv.GetProto(cfgKVKey, o.cfg); err != nil && !errors.Is(err, akcore.ErrNotFound) {
		return fmt.Errorf("retrieving config: %w", err)
	}
	if o.cfg.TrackCount == 0 {
		o.cfg.TrackCount = 5
	}

	defer o.writeCfg()

	o.handleRequests(ctx)

	return ctx.Err()
}

//go:embed web.zip
var overlayHTML []byte

func (o *Overlay) writeCfg() {
	o.lock.Lock()
	defer o.lock.Unlock()
	if err := o.kv.SetProto(cfgKVKey, o.cfg); err != nil {
		o.log.Error("writing config", "error", err.Error())
	}
}

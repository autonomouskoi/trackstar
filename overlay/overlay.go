package overlay

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/storage/kv"
	"github.com/autonomouskoi/akcore/web/webutil"
)

const (
	EnvLocalContentPath = "AK_CONTENT_TRACKSTAR_OVERLAY"
)

var (
	cfgKVKey = []byte("config")
)

func init() {
	manifest := &modules.Manifest{
		Id:          "74623a194d49d3ca",
		Title:       "TS: Overlay",
		Name:        "trackstaroverlay",
		Description: "OBS Overlay for Trackstar",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "https://autonomouskoi.org/module-trackstaroverlay.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_HELP,
				Description: "Help!",
			},
			{
				Path:        "/m/trackstaroverlay/",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_OBS_OVERLAY,
				Description: "OBS Overlay",
			},
			{
				Path:        "/m/trackstaroverlay/embed_ctrl.js",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_EMBED_CONTROL,
				Description: "Overlay Customization",
			},
			{
				Path:        "/m/trackstaroverlay/ui.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_CONTROL_PAGE,
				Description: "Overlay Customization",
			},
		},
	}
	modules.Register(manifest, &Overlay{})
}

type Overlay struct {
	http.Handler
	modutil.ModuleBase
	bus  *bus.Bus
	lock sync.Mutex
	cfg  *Config
	kv   *kv.KVPrefix
}

func (o *Overlay) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	o.Log = deps.Log
	o.bus = deps.Bus
	o.kv = &deps.KV

	o.cfg = &Config{}
	if err := o.kv.GetProto(cfgKVKey, o.cfg); err != nil && !errors.Is(err, akcore.ErrNotFound) {
		return fmt.Errorf("retrieving config: %w", err)
	}
	defer o.writeCfg()

	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, overlayHTML)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/custom-css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Content-Length", strconv.Itoa(len(o.cfg.CustomCss)))
		w.Header().Set("Cache-Control", "no-store")
		io.Copy(w, strings.NewReader(o.cfg.CustomCss))
	})
	mux.Handle("/", http.FileServer(fs))
	o.Handler = mux

	o.handleRequests(ctx)

	return ctx.Err()
}

//go:embed web.zip
var overlayHTML []byte

func (o *Overlay) writeCfg() {
	o.lock.Lock()
	defer o.lock.Unlock()
	if err := o.kv.SetProto(cfgKVKey, o.cfg); err != nil {
		o.Log.Error("writing config", "error", err.Error())
	}
}

//go:embed icon.svg
var icon []byte

func (*Overlay) Icon() ([]byte, string, error) {
	return icon, "image/svg+xml", nil
}

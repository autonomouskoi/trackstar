package overlay

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/storage/kv"
	"github.com/autonomouskoi/akcore/web/webutil"
	"github.com/autonomouskoi/trackstar"
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

	if os.Getenv("TRACKSTAR_OVERLAY_DEMO") != "" {
		go func() {
			deps.Log.Info("using demo mode")
			time.Sleep(time.Second * 5)
			for i := 1; i < 5; i++ {
				b, err := proto.Marshal(&trackstar.DeckDiscovered{
					DeckId: fmt.Sprint("Deck", i),
				})
				if err != nil {
					return
				}
				deps.Bus.Send(&bus.BusMessage{
					Topic:   trackstar.BusTopic_TRACKSTAR.String(),
					Type:    int32(trackstar.MessageType_TYPE_DECK_DISCOVERED),
					Message: b,
				})
			}
			i := 0
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second * 5):
					b, err := proto.Marshal(&trackstar.TrackUpdate{
						DeckId: fmt.Sprintf("Deck%d", i%4+1),
						Track: &trackstar.Track{
							Artist: fmt.Sprint("Artist ", i),
							Title:  fmt.Sprint("Title ", i),
						},
						When: time.Now().Unix(),
					})
					if err != nil {
						deps.Log.Error("marshalling TrackUpdate proto", "error", err.Error())
						continue
					}
					deps.Bus.Send(&bus.BusMessage{
						Topic:   trackstar.BusTopic_TRACKSTAR.String(),
						Type:    int32(trackstar.MessageType_TYPE_TRACK_UPDATE),
						Message: b,
					})
					i++
				}
			}
		}()
	}

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

package trackstar

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"

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
	manfiest := &modules.Manifest{
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
	modules.Register(manfiest, &Trackstar{})
}

//go:embed web.zip
var webZip []byte

type Trackstar struct {
	http.Handler
	updates []*TrackUpdate
}

func (ts *Trackstar) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	fs, err := webutil.ZipOrEnvPath(EnvLocalContentPath, webZip)
	if err != nil {
		return fmt.Errorf("get web FS %w", err)
	}
	ts.Handler = http.StripPrefix("/m/trackstar", http.FileServer(fs))

	in := make(chan *bus.BusMessage, 16)
	defer func() {
		deps.Bus.Unsubscribe(BusTopic_TRACKSTAR.String(), in)
		bus.Drain(in)
	}()
	deps.Bus.Subscribe(BusTopic_TRACKSTAR.String(), in)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-in:
			switch msg.Type {
			case int32(MessageType_TYPE_TRACK_UPDATE):
				tu := &TrackUpdate{}
				if err := proto.Unmarshal(msg.Message, tu); err != nil {
					deps.Log.Error("unmarshalling TrackUpdate", "error", err.Error())
					continue
				}
				ts.updates = append(ts.updates, tu)
			}
		}
	}
}

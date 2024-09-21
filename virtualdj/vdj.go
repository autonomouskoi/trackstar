package virtualdj

import (
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/util/fsutil"
	"github.com/autonomouskoi/trackstar"
)

const (
	vdjLinePrefix = "#EXTVDJ:"
)

func init() {
	manifest := &modules.Manifest{
		Id:          "8bd0166d36fec0c5",
		Name:        "trackstarvirtualdj",
		Description: "Retrieve real-time track information from VirtualDJ by reading History files",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "https://autonomouskoi.org/module-trackstarvirtualdj.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_HELP,
				Description: "Help!",
			},
		},
	}
	modules.Register(manifest, &VirtualDJ{})
}

type VirtualDJ struct {
	log *slog.Logger
	bus *bus.Bus
}

func (vdj *VirtualDJ) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	vdj.log = deps.Log
	vdj.bus = deps.Bus

	return vdj.handleVDJ(ctx)
}

type entry struct {
	Artist string `xml:"artist"`
	Title  string `xml:"title"`
	Time   string `xml:"time"`
}

func (vdj *VirtualDJ) handleVDJ(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	const timeFmt = "15:04"
	start, _ := time.Parse(timeFmt, time.Now().Format(timeFmt))

	historyDir, err := getHistoryPath()
	if err != nil {
		return fmt.Errorf("getting history path: %w", err)
	}
	playlist := filepath.Join(historyDir, time.Now().Format("2006-01-02")+".m3u")

	loggedWaiting := false
	for {
		if ctx.Err() != nil {
			return nil
		}
		_, err := os.Stat(playlist)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			vdj.log.Info("found playlist", "path", playlist)
			break
		}
		if !loggedWaiting {
			vdj.log.Info("waiting for history playlist", "path", playlist)
			loggedWaiting = true
		}
		time.Sleep(time.Second * 5)
	}

	rc, err := fsutil.NewPollingFileAppendReader(
		playlist,
		time.Second*5,
	)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		rc.Close()
	}()

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, vdjLinePrefix) {
			continue
		}
		e, err := parseLine(line)
		if err != nil {
			return err
		}
		trackTime, err := time.Parse(timeFmt, e.Time)
		if err != nil {
			vdj.log.Error("parsing track time", "time", e.Time, "error", err.Error())
			continue
		}
		if trackTime.Before(start) {
			vdj.log.Debug("skipping track", "time", e.Time, "artist", e.Artist, "title", e.Title)
			continue
		}
		vdj.log.Debug("sending track", "artist", e.Artist, "title", e.Title)
		b, _ := proto.Marshal(&trackstar.SubmitTrackRequest{
			TrackUpdate: &trackstar.TrackUpdate{
				Track: &trackstar.Track{
					Artist: e.Artist,
					Title:  e.Title,
				},
				When: time.Now().Unix(),
			}})
		vdj.bus.Send(&bus.BusMessage{
			Topic:   trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
			Type:    int32(trackstar.MessageTypeRequest_SUBMIT_TRACK_REQ),
			Message: b,
		})
	}

	return scanner.Err()
}

func parseLine(line string) (entry, error) {
	line = "<entry>" + strings.TrimPrefix(line, vdjLinePrefix) + "</entry>"
	decoder := xml.NewDecoder(strings.NewReader(line))
	decoder.Strict = false
	var e entry
	if err := decoder.Decode(&e); err != nil {
		return e, fmt.Errorf("parsing XML %q: %w", line, err)
	}
	return e, nil
}

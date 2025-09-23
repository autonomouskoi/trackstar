package rekordboxdb

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	_ "github.com/xeodou/go-sqlcipher"
	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/akcore/svc/log"
	trackstar "github.com/autonomouskoi/trackstar/pb"
)

func init() {
	manifest := &modules.Manifest{
		Id:          "ce1f91f7dc0fa32c",
		Title:       "TS: Rekordbox",
		Name:        "trackstarrekordboxdb",
		Description: "Retrieve real-time track information directly from the Rekordbox database. There's a delay configured in Rekordbox Preferences -> Advanced -> Browse -> Playback time setting",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "https://autonomouskoi.org/module-trackstarrekordboxdb.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_HELP,
				Description: "Help!",
			},
		},
	}
	modules.Register(manifest, &RekordboxDB{})
}

type RekordboxDB struct {
	db  db
	bus *bus.Bus
	log log.Logger
}

func (rbdb *RekordboxDB) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	rbdb.log = deps.Log
	rbdb.bus = deps.Bus

	var err error
	rbdb.db, err = newDB()
	if err != nil {
		return fmt.Errorf("opening Rekordbox database: %w", err)
	}
	defer rbdb.db.Close()

	err = rbdb.handleRB(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (rbdb *RekordboxDB) handleRB(ctx context.Context) error {
	// on a fresh database, this fails. Keep trying until they play a song
	var err error
	var historyID string
	var historyEntries []*dbSongHistoryEntry
	sleepTime := time.Duration(0)
	for {
		if ctx.Err() != nil {
			return nil
		}
		time.Sleep(sleepTime)
		sleepTime = time.Second * 5 // don't spam the logs
		historyID, err = rbdb.db.latestHistoryID(ctx)
		if err != nil {
			rbdb.log.Debug("no history yet")
			continue
		}
		historyEntries, err = rbdb.db.GetSongHistory(ctx, historyID)
		if err != nil {
			rbdb.log.Debug("no song history yet")
			continue
		}
		break
	}

	lastHistoryCheck := time.Now()
	var lastTrackNo int64
	if len(historyEntries) > 0 {
		lastTrackNo = historyEntries[len(historyEntries)-1].TrackNo
	}

	for ctx.Err() == nil {
		time.Sleep(time.Second / 4)
		if time.Since(lastHistoryCheck) > time.Second*5 {
			historyID, err = rbdb.db.latestHistoryID(ctx)
			if err != nil {
				return fmt.Errorf("getting latest history ID: %w", err)
			}
			lastHistoryCheck = time.Now()
		}

		historyEntries, err = rbdb.db.GetSongHistory(ctx, historyID)
		if err != nil {
			return fmt.Errorf("getting history: %w", err)
		}
		if len(historyEntries) == 0 {
			continue
		}

		latestEntry := historyEntries[len(historyEntries)-1]
		if trackNo := latestEntry.TrackNo; trackNo != lastTrackNo {
			track, err := rbdb.db.GetTrack(ctx, latestEntry.ContentID)
			if err != nil {
				rbdb.log.Error("getting track", "content_id", latestEntry.ContentID, "error", err.Error())
				continue
			}
			b, err := proto.Marshal(&trackstar.SubmitTrackRequest{
				TrackUpdate: &trackstar.TrackUpdate{
					DeckId: "Rekordbox",
					Track:  track,
					When:   time.Now().Unix(),
				}})
			if err != nil {
				rbdb.log.Error("marshalling TrackUpdate", "error", err.Error())
				continue
			}
			rbdb.log.Debug("sending track", "track", track)
			subCtx, cancel := context.WithTimeout(ctx, time.Second)
			_ = rbdb.bus.WaitForReply(subCtx, &bus.BusMessage{
				Topic:   trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
				Type:    int32(trackstar.MessageTypeRequest_SUBMIT_TRACK_REQ),
				Message: b,
			})
			cancel()
			lastTrackNo = trackNo
		}
	}
	return nil
}

//go:embed icon.svg
var icon []byte

func (*RekordboxDB) Icon() ([]byte, string, error) {
	return icon, "image/svg+xml", nil
}

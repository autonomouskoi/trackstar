package mixxx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/xeodou/go-sqlcipher"

	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/trackstar"
)

func init() {
	manifest := &modules.Manifest{
		Id:          "2d1b8258de88edc7",
		Name:        "trackstarmixxx",
		Description: "Retrieve real-time track information directly from the Mixxx History",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "https://autonomouskoi.org/module-trackstarmixxx.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_HELP,
				Description: "Help!",
			},
		},
	}
	modules.Register(manifest, &Mixxx{})
}

type Mixxx struct {
	modutil.ModuleBase
	db  *sqlx.DB
	bus *bus.Bus
	Log *slog.Logger
}

func (m *Mixxx) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	m.Log = deps.Log
	m.bus = deps.Bus

	dbPath, err := getDBPath()
	if err != nil {
		return fmt.Errorf("getting Mixxx DB path: %w", err)
	}

	m.db, err = sqlx.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath))
	if err != nil {
		return fmt.Errorf("opening db %s: %w", dbPath, err)
	}
	defer m.db.Close()

	currentTrackID, err := m.getLatest(ctx)
	if err != nil {
		return fmt.Errorf("getting latest track: %w", err)
	}

	for {
		time.Sleep(time.Second / 4)
		if ctx.Err() != nil {
			break
		}
		newestID, err := m.getLatest(ctx)
		if err != nil {
			return fmt.Errorf("getting latest track: %w", err)
		}
		if currentTrackID == newestID {
			continue
		}
		currentTrackID = newestID
		track, err := m.getTrack(ctx, newestID)
		if err != nil {
			return fmt.Errorf("getting track metadata: %w", err)
		}
		msg := &bus.BusMessage{
			Topic: trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
			Type:  int32(trackstar.MessageTypeRequest_SUBMIT_TRACK_REQ),
		}
		m.MarshalMessage(msg, &trackstar.SubmitTrackRequest{
			TrackUpdate: &trackstar.TrackUpdate{
				DeckId: "Mixxx",
				Track:  track,
				When:   time.Now().Unix(),
			},
		})
		m.Log.Debug("sending track", "artist", track.Artist, "title", track.Title)
		subCtx, cancel := context.WithTimeout(ctx, time.Second)
		_ = m.bus.WaitForReply(subCtx, msg)
		cancel()
	}

	return nil
}

func (m *Mixxx) getLatest(ctx context.Context) (int, error) {
	v := struct {
		ID int `db:"id"`
	}{}
	err := m.db.GetContext(ctx, &v, `
SELECT track_id
	FROM PlaylistTracks
	ORDER BY pl_datetime_added DESC
	LIMIT 1`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // no history yet. Weird but ok
		}
		return 0, err
	}
	return v.ID, nil
}

func (m *Mixxx) getTrack(ctx context.Context, id int) (*trackstar.Track, error) {
	t := struct {
		Artist sql.NullString `db:"artist"`
		Title  sql.NullString `db:"title"`
	}{}
	err := m.db.GetContext(ctx, &t, m.db.Rebind(`
SELECT artist, title
	FROM library
	WHERE id = ?
	LIMIT 1
`), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // really weird, but ok
		}
		return nil, err
	}
	return &trackstar.Track{
		Artist: t.Artist.String,
		Title:  t.Title.String,
	}, nil
}

func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user homedir: %w", err)
	}
	return filepath.Join(homeDir, "Library", "Containers", "org.mixxx.mixxx", "Data", "Library", "Application Support", "Mixxx", "mixxxdb.sqlite"), nil
}

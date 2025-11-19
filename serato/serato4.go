package serato

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/xeodou/go-sqlcipher" // ideally it's sqlite directly, but save a dependency

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/svc/log"
	trackstar "github.com/autonomouskoi/trackstar/pb"
)

type Serato4 struct {
	log   log.Logger
	db    *sqlx.DB
	maxID int
}

func New(log log.Logger) (*Serato4, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", akcore.ErrNotFound, err)
	}
	dsn := fmt.Sprintf("file:%s?mode=ro", dbPath)
	log.Debug("connecting to database", "dsn", dsn)
	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("%w: opening database: %w", akcore.ErrNotFound, err)
	}

	s4 := &Serato4{
		log: log,
		db:  db,
	}

	err = db.Get(&s4.maxID, "SELECT max(id) FROM history_entry")
	if err != nil {
		if err := db.Close(); err != nil {
			log.Error("closing serato DB", "dsn", dsn, "error", err.Error())
		}
		return nil, fmt.Errorf("%w: querying history: %w", akcore.ErrNotFound, err)
	}
	// if we've gotten this far there's a functional 4.0 database
	log.Debug("selected from history_entry", "max_id", s4.maxID)

	return s4, nil
}

func (s4 *Serato4) Go(ctx context.Context, handler func(tu *trackstar.TrackUpdate)) error {
	defer func() {
		if err := s4.db.Close(); err != nil {
			s4.log.Error("closing serato DB", "error", err.Error())
		}
	}()
	for {
		select {
		case <-ctx.Done():
			s4.log.Debug("got cancellation")
			return nil
		case <-time.After(time.Second):
			if err := s4.readTrack(ctx, handler); err != nil {
				s4.log.Error("reading track", "error", err.Error())
			}
		}
	}
}

type HistoryEntry struct {
	ID     int    `db:"id"`
	Artist string `db:"artist"`
	Title  string `db:"name"`
	Device string `db:"device"`
	Deck   string `db:"deck"`
}

func (s4 *Serato4) readTrack(ctx context.Context, handler func(tu *trackstar.TrackUpdate)) error {
	var entry HistoryEntry
	err := s4.db.GetContext(ctx, &entry, `
SELECT id, artist, name, device, deck FROM history_entry
	WHERE id > ?
		ORDER BY id ASC
		LIMIT 1
`, s4.maxID)
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, context.Canceled) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting next track: %w", err)
	}
	s4.maxID = entry.ID
	handler(&trackstar.TrackUpdate{
		DeckId: entry.Device + "/" + entry.Deck,
		When:   time.Now().Unix(),
		Track: &trackstar.Track{
			Artist: entry.Artist,
			Title:  entry.Title,
		},
	})
	return nil
}

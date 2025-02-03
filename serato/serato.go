package serato

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/trackstar"
)

var zeroTrack Track

func init() {
	manifest := &modules.Manifest{
		Id:          "9be27608da1d141b",
		Title:       "TS: Serato",
		Name:        "trackstarserato",
		Description: "Retrieve real-time track information from Serato by reading session files",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "https://autonomouskoi.org/module-trackstarserato.html",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_HELP,
				Description: "Help!",
			},
		},
	}
	modules.Register(manifest, &Serato{})
}

type Serato struct {
	modutil.ModuleBase
	bus *bus.Bus
}

func (s *Serato) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	s.Log = deps.Log
	s.bus = deps.Bus

	sessionDirPath, err := getSessionsPath()
	if err != nil {
		return fmt.Errorf("getting session path: %w", err)
	}

	sf := newSessionFile(sessionDirPath)
	sf.handleTrack = s.handleTrack
	track, err := sf.discoverFile(ctx)
	if err != nil {
		return fmt.Errorf("discovering session file: %w", err)
	}
	s.handleTrack(track)

	return sf.watch(ctx)
}

func (s *Serato) handleTrack(t Track) {
	msg := &bus.BusMessage{
		Topic: trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
		Type:  int32(trackstar.MessageTypeRequest_SUBMIT_TRACK_REQ),
	}
	str := &trackstar.SubmitTrackRequest{
		TrackUpdate: &trackstar.TrackUpdate{
			When: time.Now().Unix(),
			Track: &trackstar.Track{
				Artist: t.Artist,
				Title:  t.Title,
			},
		},
	}
	s.MarshalMessage(msg, str)
	s.bus.Send(msg)
}

type sessionFile struct {
	dir         string
	name        string
	mod         time.Time
	offset      int64
	handleTrack func(Track)
}

func newSessionFile(dir string) *sessionFile {
	return &sessionFile{
		dir: dir,
		mod: time.Now(),
	}
}

func (sf *sessionFile) discoverFile(ctx context.Context) (Track, error) {
	for {
		des, err := os.ReadDir(sf.dir)
		if err != nil {
			return zeroTrack, fmt.Errorf("reading session directory: %w", err)
		}
		for _, de := range des {
			if !strings.HasSuffix(de.Name(), ".session") {
				continue
			}
			if de.IsDir() {
				continue
			}
			fi, err := de.Info()
			if err != nil {
				return zeroTrack, fmt.Errorf("getting info for %s: %w", de.Name(), err)
			}
			if !fi.ModTime().After(sf.mod) {
				continue
			}
			track, err := sf.getLatestTrack(de.Name())
			if err != nil {
				return zeroTrack, fmt.Errorf("finding initial newest track in %s: %w", de.Name(), err)
			}
			if track != zeroTrack {
				sf.mod = fi.ModTime()
				sf.name = de.Name()
				return track, nil
			}
		}
		select {
		case <-ctx.Done():
			return zeroTrack, ctx.Err()
		case <-time.After(time.Second * 5):
		}
	}
}

func (sf *sessionFile) watch(ctx context.Context) error {
	for {
		sessionPath := filepath.Join(sf.dir, sf.name)
		stat, err := os.Stat(sessionPath)
		if err != nil {
			return fmt.Errorf("stating session %s: %w", sf.name, err)
		}
		if stat.ModTime().After(sf.mod) {
			track, err := sf.getLatestTrack(sf.name)
			if err != nil {
				return fmt.Errorf("getting latest track from %s: %w", sf.name, err)
			}
			if track != zeroTrack {
				sf.handleTrack(track)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 5):
		}
	}
}

func (sf *sessionFile) getLatestTrack(fileName string) (Track, error) {
	var latest Track
	infh, err := os.Open(filepath.Join(sf.dir, fileName))
	if err != nil {
		return latest, fmt.Errorf("opening session: %w", err)
	}
	defer infh.Close()
	if _, err := infh.Seek(int64(sf.offset), io.SeekStart); err != nil {
		return latest, fmt.Errorf("seeking: %w", err)
	}

	latest, err = getLatestTrackAfter(infh, sf.mod)
	if err != nil {
		return latest, fmt.Errorf("reading session: %w", err)
	}
	if latest != zeroTrack {
		sf.offset, err = infh.Seek(0, io.SeekCurrent)
		if err != nil {
			return latest, fmt.Errorf("getting offset: %w", err)
		}
	}
	return latest, nil
}

func getLatestTrackAfter(r io.Reader, after time.Time) (Track, error) {
	var latest Track
	err := ReadSession(r, func(t Track) {
		if t.When.After(after) {
			latest = t
		}
	})
	if err != nil {
		return latest, fmt.Errorf("reading session: %w", err)
	}
	return latest, nil
}

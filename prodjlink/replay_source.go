package prodjlink

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"go.evanpurkhiser.com/prolink"
)

type ReplaySource struct {
	lock   sync.Mutex
	path   string
	tracks map[uint64]*prolink.Track
}

func newReplaySource(path string) *ReplaySource {
	return &ReplaySource{
		path:   path,
		tracks: map[uint64]*prolink.Track{},
	}
}

func (s *ReplaySource) recv(ctx context.Context, out chan<- *prolink.CDJStatus) error {
	defer close(out)
	infh, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("opening %s: %w", s.path, err)
	}
	defer infh.Close()

	jd := json.NewDecoder(infh)

	for {
		if err := ctx.Err(); err != nil {
			break
		}
		time.Sleep(time.Second / 100)
		event := struct {
			Event  string             `json:"event"`
			Status *prolink.CDJStatus `json:"status"`
			Track  *prolink.Track     `json:"track"`
		}{}
		err := jd.Decode(&event)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("decoding: %w", err)
		}

		if event.Status.TrackKey() != nil {
			s.lock.Lock()
			cacheKey := trackKey(event.Status.TrackKey())
			if _, present := s.tracks[cacheKey]; !present && event.Track != nil {
				s.tracks[cacheKey] = event.Track
			}
			s.lock.Unlock()
		}

		out <- event.Status
	}
	return nil
}

func (s *ReplaySource) getTrack(tk *prolink.TrackKey) *prolink.Track {
	if tk == nil || tk.TrackID == 0 {
		return noTrack
	}
	cacheKey := trackKey(tk)
	s.lock.Lock()
	defer s.lock.Unlock()
	if track, present := s.tracks[cacheKey]; present {
		return track
	}
	return noTrack
}

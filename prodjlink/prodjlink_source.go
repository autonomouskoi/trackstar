package prodjlink

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/autonomouskoi/datastruct/ttlcache"
	"go.evanpurkhiser.com/prolink"
	"go.evanpurkhiser.com/prolink/mixstatus"

	"github.com/autonomouskoi/akcore/svc/log"
)

var noTrack = &prolink.Track{
	Artist: "ID",
	Title:  "ID",
}

// create a uint64 from a TrackKey by packing byte fields. This depends on
// endianness and should be marshalled with care
func trackKey(tk *prolink.TrackKey) uint64 {
	return (uint64(tk.TrackID) << 24) |
		(uint64(tk.Slot) << 16) |
		(uint64(tk.Type) << 8) |
		uint64(tk.DeviceID)
}

type ProDJLinkSource struct {
	lock       sync.Mutex
	trackCache ttlcache.Cache[uint64, *prolink.Track]
	dj         *prolink.CDJStatusMonitor
	rbdb       *prolink.RemoteDB
}

func newProDJLinkSource(ctx context.Context, log log.Logger) (*ProDJLinkSource, error) {
	prolink.Log = pdjLogger{log}

	network, err := prolink.Connect()
	if err != nil {
		return nil, fmt.Errorf("connecting: %w", err)
	}
	if err := network.AutoConfigure(5 * time.Second); err != nil {
		return nil, fmt.Errorf("autoconfiguring: %w", err)
	}

	dm := network.DeviceManager()
	dm.OnDeviceAdded("", prolink.DeviceListenerFunc(func(d *prolink.Device) {
		log.Info("device added", "device", d)
	}))
	dm.OnDeviceRemoved("", prolink.DeviceListenerFunc(func(d *prolink.Device) {
		log.Info("device removed", "device", d)
	}))

	return &ProDJLinkSource{
		dj:         network.CDJStatusMonitor(),
		rbdb:       network.RemoteDB(),
		trackCache: ttlcache.New[uint64, *prolink.Track](ctx, time.Minute*5, time.Minute*5),
	}, nil
}

func (s *ProDJLinkSource) recv(ctx context.Context, out chan<- *prolink.CDJStatus) error {
	s.dj.AddStatusHandler(mixstatus.NewProcessor(
		mixstatus.Config{
			AllowedInterruptBeats: 8,
			BeatsUntilReported:    16,
			TimeBetweenSets:       time.Second * 10,
		},
		mixstatus.HandlerFunc(func(e mixstatus.Event, c *prolink.CDJStatus) {
			out <- c
		}),
	))
	<-ctx.Done()
	return nil
}

func (s *ProDJLinkSource) getTrack(tk *prolink.TrackKey) *prolink.Track {
	if tk == nil || tk.TrackID == 0 {
		return noTrack
	}
	cacheKey := trackKey(tk)
	s.lock.Lock()
	defer s.lock.Unlock()
	track, present := s.trackCache.Get(cacheKey)
	if present {
		return track
	}

	track, err := s.rbdb.GetTrack(tk)
	if err != nil {
		return noTrack
	}
	s.trackCache.Set(cacheKey, track)
	return track
}

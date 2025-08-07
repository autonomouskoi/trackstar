// Package prodjlink provides track data to Trackstar from Pioneer/AlphaTheta
// devices that support the Pro DJ Link protocol.
package prodjlink

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.evanpurkhiser.com/prolink"

	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/akcore/modules/modutil"
	"github.com/autonomouskoi/trackstar"
)

func init() {
	manifest := &modules.Manifest{
		Id:          "371b0b412d302cfd",
		Title:       "TS: Pro DJ Link (Experimental)",
		Name:        "trackstarprodjlink",
		Description: "Retrieve real-time track information from Pro DJ Link capable Pioneer/AlphaTheta devices",
		WebPaths:    []*modules.ManifestWebPath{},
	}
	modules.Register(manifest, &ProDJLink{})
}

type ProDJLink struct {
	//http.Handler
	modutil.ModuleBase
	bus *bus.Bus
}

type Source interface {
	recv(context.Context, chan<- *prolink.CDJStatus) error
	getTrack(*prolink.TrackKey) *prolink.Track
}

func (pl *ProDJLink) Start(ctx context.Context, deps *modutil.ModuleDeps) error {
	pl.Log = deps.Log
	pl.bus = deps.Bus

	var source Source
	if v := os.Getenv("PRODJLINK_REPLAY_PATH"); v != "" {
		source = newReplaySource("/tmp/blarg")
	} else {
		var err error
		source, err = newProDJLinkSource(ctx, pl.Log)
		if err != nil {
			return fmt.Errorf("creating Pro DJ Link source: %w", err)
		}
	}
	in := make(chan *prolink.CDJStatus, 16)
	pl.Go(func() error { return source.recv(ctx, in) })

	players := map[string]DeviceStatus{}

	msg := &bus.BusMessage{
		Topic: trackstar.BusTopic_TRACKSTAR_REQUEST.String(),
		Type:  int32(trackstar.MessageTypeRequest_SUBMIT_TRACK_REQ),
	}
	for status := range in {
		deckID := DeckID(status)
		newStatus := DeviceStatus{
			Status: status,
		}
		status, ok := players[deckID]
		if !ok {
			players[deckID] = newStatus
			pl.Log.Debug("adding deck", "deck_id", deckID)
			continue
		}
		players[deckID] = newStatus
		if status.Status.PlayState != prolink.PlayStatePlaying && newStatus.Status.PlayState == prolink.PlayStatePlaying {
			track := source.getTrack(status.Status.TrackKey())
			pl.MarshalMessage(msg, &trackstar.SubmitTrackRequest{
				TrackUpdate: &trackstar.TrackUpdate{
					DeckId: deckID,
					Track: &trackstar.Track{
						Artist: track.Artist,
						Title:  track.Title,
					},
					When: time.Now().Unix(),
				},
			})
			if msg.Error != nil {
				pl.Log.Error("marshalling track", "error", msg.Error.UserMessage)
				continue
			}
			pl.Log.Debug("sending track", "artist", track.Artist, "title", track.Title)
			pl.bus.Send(msg)
		}
	}

	return pl.Wait()
}

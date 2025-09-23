package trackstar

import (
	"context"
	"time"

	"golang.org/x/exp/rand"

	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/trackstar/pb"
)

var demoArtists = []string{
	"A.B.C",
	"Agressor Dunx",
	"Juan V",
	"Visual",
	"AC Breads",
	"Nausea",
	"Zybra",
	"Crack Sum Umpire",
	"Teddy Stufferz",
	"DJ Hazmat",
}

var demoTitles = []string{
	"Grampino",
	"A Very Long Song Title For The Sake of Nox's Testing",
	"Iraqis",
	"Ma'am Eating Wizard Beardo",
	"Laser Bean",
	"Hard Noids",
	"Crab Thug",
	"Roti Poti",
}

func (ts *Trackstar) demoMode() {
	if ts.cfg.DemoDelaySeconds == 0 {
		return
	}

	rand := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

	interval := time.Second * time.Duration(ts.cfg.DemoDelaySeconds)
	ctx, demoCancel := context.WithCancel(context.Background())
	ts.demoCancel = demoCancel
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				tu := &pb.TrackUpdate{
					When:   time.Now().Unix(),
					DeckId: "Demo",
					Track:  &pb.Track{},
				}
				str := &pb.SubmitTrackRequest{
					TrackUpdate: tu,
				}
				tu.Track.Artist = demoArtists[rand.Intn(len(demoArtists))]
				tu.Track.Title = demoTitles[rand.Intn(len(demoTitles))]
				msg := &bus.BusMessage{
					Topic: pb.BusTopic_TRACKSTAR_REQUEST.String(),
					Type:  int32(pb.MessageTypeRequest_SUBMIT_TRACK_REQ),
				}
				ts.MarshalMessage(msg, str)
				_ = ts.bus.WaitForReply(ctx, msg)
			}
		}

	}()
}

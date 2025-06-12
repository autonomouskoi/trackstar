package trackstar

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore"
	"github.com/autonomouskoi/akcore/bus"
)

func (ts *Trackstar) handleRequests(ctx context.Context) error {
	ts.bus.HandleTypes(ctx, BusTopic_TRACKSTAR_REQUEST.String(), 8,
		map[int32]bus.MessageHandler{
			int32(MessageTypeRequest_GET_TRACK_REQ):     ts.handleGetTrackRequest,
			int32(MessageTypeRequest_SUBMIT_TRACK_REQ):  ts.handleRequestSubmitTrack,
			int32(MessageTypeRequest_CONFIG_GET_REQ):    ts.handleRequestConfigGet,
			int32(MessageTypeRequest_GET_SESSION_REQ):   ts.handleRequestGetSession,
			int32(MessageTypeRequest_TAG_TRACK_REQ):     ts.handleRequestTagTrack,
			int32(MessageTypeRequest_LIST_SESSIONS_REQ): ts.handleRequestListSessions,
		},
		nil,
	)
	return nil
}

func (ts *Trackstar) handleGetTrackRequest(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	gtr := &GetTrackRequest{}
	if reply.Error = ts.UnmarshalMessage(msg, gtr); reply.Error != nil {
		return reply
	}
	ts.lock.Lock()
	defer ts.lock.Unlock()
	if len(ts.session.Tracks) == 0 {
		reply.Error = &bus.Error{
			Code:   int32(bus.CommonErrorCode_NOT_FOUND),
			Detail: proto.String("no tracks"),
		}
		return reply
	}
	when := time.Now().Add(-time.Second * time.Duration(gtr.DeltaSeconds)).Unix()
	gtResp := &GetTrackResponse{
		TrackUpdate: ts.session.Tracks[0],
	}
	for _, tu := range ts.session.Tracks[1:] {
		if tu.When > when {
			break
		}
		gtResp.TrackUpdate = tu
	}
	ts.MarshalMessage(reply, gtResp)
	return reply
}

var bracketRE = regexp.MustCompile(`\[.*\]`)
var multispaceRE = regexp.MustCompile(`\s{2,}`)

func (ts *Trackstar) mungeTrackUpdate(tu *TrackUpdate) {
	for match, replace := range ts.cfg.TrackReplacements {
		if strings.TrimSpace(match) == "" {
			continue
		}
		if strings.Contains(tu.Track.Artist, match) || strings.Contains(tu.Track.Title, match) {
			tu.Track.Artist = replace.Artist
			tu.Track.Title = replace.Title
			return
		}
	}
	if ts.cfg.ClearBracketedText {
		tu.Track.Artist = bracketRE.ReplaceAllString(tu.Track.Artist, " ")
		tu.Track.Artist = multispaceRE.ReplaceAllString(tu.Track.Artist, " ")
		tu.Track.Title = bracketRE.ReplaceAllString(tu.Track.Title, " ")
		tu.Track.Title = multispaceRE.ReplaceAllString(tu.Track.Title, " ")
	}
	tu.Track.Artist = strings.TrimSpace(tu.Track.Artist)
	tu.Track.Title = strings.TrimSpace(tu.Track.Title)
}

func (ts *Trackstar) handleRequestSubmitTrack(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.Topic,
		Type:  msg.Type + 1,
	}
	str := &SubmitTrackRequest{}
	if reply.Error = ts.UnmarshalMessage(msg, str); reply.Error != nil {
		return reply
	}
	ts.mungeTrackUpdate(str.TrackUpdate)

	go func() {
		time.Sleep(time.Second * time.Duration(ts.cfg.TrackDelaySeconds))
		ts.lock.Lock()
		ts.session.Tracks = append(ts.session.Tracks, str.TrackUpdate)
		ts.lock.Unlock()
		if ts.cfg.SaveSessions {
			if err := ts.saveSession(ts.session); err != nil {
				ts.Log.Error("saving session", "error", err.Error())
			}
		}

		tuMsg := &bus.BusMessage{
			Topic: BusTopic_TRACKSTAR_EVENT.String(),
			Type:  int32(MessageTypeEvent_TRACK_UPDATE),
		}
		tuMsg.Message, _ = proto.Marshal(str.TrackUpdate)
		ts.Log.Debug("sending track", "deck_id", str.TrackUpdate.DeckId,
			"artist", str.TrackUpdate.Track.Artist,
			"title", str.TrackUpdate.Track.Title,
		)
		ts.bus.Send(tuMsg)
	}()

	ts.MarshalMessage(reply, &SubmitTrackResponse{})
	return reply
}

func (ts *Trackstar) handleRequestConfigGet(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	ts.lock.Lock()
	ts.MarshalMessage(reply, &ConfigGetResponse{
		Config: ts.cfg,
	})
	ts.lock.Unlock()
	return reply
}

func (ts *Trackstar) handleRequestGetSession(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	req := &GetSessionRequest{}
	if reply.Error = ts.UnmarshalMessage(msg, req); reply.Error != nil {
		return reply
	}
	ts.lock.Lock()
	defer ts.lock.Unlock()
	session := ts.session
	if req.Session != 0 && req.Session != ts.session.Started {
		session = &Session{}
		err := ts.kv.GetProto([]byte(fmt.Sprintf("%s%d", sessionPrefix, req.Session)), session)
		if err != nil {
			if errors.Is(err, akcore.ErrNotFound) {
				reply.Error = &bus.Error{Code: int32(bus.CommonErrorCode_NOT_FOUND)}
			} else {
				reply.Error = &bus.Error{Detail: proto.String(err.Error())}
			}
			return reply
		}
	}
	ts.MarshalMessage(reply, &GetSessionResponse{
		Session: session,
	})
	return reply
}

func (ts *Trackstar) handleRequestTagTrack(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	ts.lock.Lock()
	defer ts.lock.Unlock()
	if len(ts.session.Tracks) == 0 {
		reply.Error = &bus.Error{Code: int32(bus.CommonErrorCode_NOT_FOUND)}
		return reply
	}
	ttr := &TagTrackRequest{}
	if reply.Error = ts.UnmarshalMessage(msg, ttr); reply.Error != nil {
		return reply
	}
	if !slices.ContainsFunc(ts.cfg.Tags, func(t *TrackTagConfig) bool { return t.Tag == ttr.Tag.GetTag() }) {
		reply.Error = &bus.Error{Code: int32(bus.CommonErrorCode_NOT_FOUND)}
		return reply
	}
	current := ts.session.Tracks[len(ts.session.Tracks)-1]
	current.Tags = append(current.Tags, ttr.GetTag())
	eventMsg := &bus.BusMessage{
		Topic: BusTopic_TRACKSTAR_EVENT.String(),
		Type:  int32(MessageTypeEvent_SESSION_UPDATE),
	}
	reply.Error = ts.UnmarshalMessage(eventMsg, &TracklogUpdateEvent{})
	ts.bus.Send(eventMsg)
	return reply
}

func (ts *Trackstar) handleRequestListSessions(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	keys, err := ts.kv.List([]byte(sessionPrefix))
	if err != nil {
		reply.Error = &bus.Error{
			Detail: proto.String("listing keys: " + err.Error()),
		}
		return reply
	}
	resp := &ListSessionsResponse{}
	for _, key := range keys {
		idStr := strings.TrimPrefix(string(key), sessionPrefix)
		sessionID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ts.Log.Error("bad session key", "key", string(key))
			continue
		}
		resp.Sessions = append(resp.Sessions, sessionID)
	}
	ts.lock.Lock()
	if len(resp.Sessions) == 0 || resp.Sessions[len(resp.Sessions)-1] != ts.session.Started {
		resp.Sessions = append(resp.Sessions, ts.session.Started)
	}
	ts.lock.Unlock()
	ts.MarshalMessage(reply, resp)
	return reply
}

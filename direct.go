package trackstar

import (
	"context"
	"time"

	"github.com/autonomouskoi/akcore/bus"
	svc "github.com/autonomouskoi/akcore/svc/pb"
)

func (ts *Trackstar) handleDirect(ctx context.Context) error {
	ts.bus.HandleTypes(ctx, moduleID, 4,
		map[int32]bus.MessageHandler{
			int32(svc.MessageType_WEBHOOK_CALL_EVENT): ts.handleDirectWebhookCall,
		},
		nil,
	)
	return nil
}

func (ts *Trackstar) handleDirectWebhookCall(msg *bus.BusMessage) *bus.BusMessage {
	wce := &svc.WebhookCallEvent{}
	if err := ts.UnmarshalMessage(msg, wce); err != nil {
		return nil
	}
	action := wce.GetParam("action")
	switch action {
	case "add_tag":
		ts.webhookCallAddTag(wce)
	default:
		ts.Log.Error("invalid webhook action", "action", action)
	}
	return nil
}

func (ts *Trackstar) webhookCallAddTag(wce *svc.WebhookCallEvent) {
	tag := wce.GetParam("tag")
	if tag == "" {
		return
	}
	msg := &bus.BusMessage{
		Topic: BusTopic_TRACKSTAR_REQUEST.String(),
		Type:  int32(MessageTypeRequest_TAG_TRACK_REQ),
	}
	ts.MarshalMessage(msg, &TagTrackRequest{
		Tag: &TrackUpdateTag{
			When:      time.Now().Unix(),
			FromId:    "0",
			FromLogin: "_webhook",
			Tag:       tag,
		},
	})
	if msg.Error != nil {
		return
	}
	ts.bus.Send(msg)
}

package trackstar

import (
	"context"
	"time"

	"github.com/autonomouskoi/akcore/bus"
)

func (ts *Trackstar) handleDirect(ctx context.Context) error {
	ts.bus.HandleTypes(ctx, moduleID, 4,
		map[int32]bus.MessageHandler{
			int32(bus.MessageTypeDirect_WEBHOOK_CALL_REQ): ts.handleDirectWebhookCall,
		},
		nil,
	)
	return nil
}

func (ts *Trackstar) handleDirectWebhookCall(msg *bus.BusMessage) *bus.BusMessage {
	wcr := &bus.WebhookCallRequest{}
	if err := ts.UnmarshalMessage(msg, wcr); err != nil {
		return nil
	}
	action := wcr.GetParam("action")
	switch action {
	case "add_tag":
		ts.webhookCallAddTag(wcr)
	default:
		ts.Log.Error("invalid webhook action", "action", action)
	}
	return nil
}

func (ts *Trackstar) webhookCallAddTag(wcr *bus.WebhookCallRequest) {
	tag := wcr.GetParam("tag")
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

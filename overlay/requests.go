package overlay

import (
	"context"

	"github.com/autonomouskoi/akcore/bus"
)

func (o *Overlay) handleRequests(ctx context.Context) {
	o.bus.HandleTypes(ctx, BusTopic_TRACKSTAR_OVERLAY_REQUEST.String(), 16,
		map[int32]bus.MessageHandler{
			int32(MessageType_GET_CONFIG_REQUEST): o.handleGetConfigRequest,
			int32(MessageType_CONFIG_SET_REQ):     o.handleConfigSetRequest,
		},
		nil)
}

func (o *Overlay) handleGetConfigRequest(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	o.lock.Lock()
	o.MarshalMessage(reply, &GetConfigResponse{
		Config: o.cfg,
	})
	o.lock.Unlock()
	return reply
}

func (o *Overlay) handleConfigSetRequest(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.Type + 1,
	}
	csr := &ConfigSetRequest{}
	if reply.Error = o.UnmarshalMessage(msg, csr); reply.Error != nil {
		return reply
	}
	o.lock.Lock()
	o.cfg = csr.Config
	defer o.writeCfg()
	o.lock.Unlock()
	o.MarshalMessage(reply, &GetConfigResponse{
		Config: o.cfg,
	})

	// also, notify the overlay
	o.bus.Send(&bus.BusMessage{
		Topic: BusTopic_TRACKSTAR_OVERLAY_EVENT.String(),
		Type:  int32(MessageType_CONFIG_UPDATED),
	})

	return reply
}

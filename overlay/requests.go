package overlay

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore/bus"
)

func (o *Overlay) handleRequests(ctx context.Context) {
	in := make(chan *bus.BusMessage, 16)
	defer func() {
		o.deps.Bus.Unsubscribe(BusTopic_TRACKSTAR_OVERLAY_REQUEST.String(), in)
		bus.Drain(in)
	}()
	o.deps.Bus.Subscribe(BusTopic_TRACKSTAR_OVERLAY_REQUEST.String(), in)
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-in:
			switch msg.Type {
			case int32(MessageType_SET_STYLE):
				o.handleSetStyle(msg)
			case int32(MessageType_GET_CONFIG_REQUEST):
				o.handleGetConfigRequest(msg)
			default:
				o.deps.Log.Error("unhandled message type", "type", msg.Type)
			}
		}
	}
}

func (o *Overlay) handleSetStyle(msg *bus.BusMessage) {
	su := &StyleUpdate{}
	if err := proto.Unmarshal(msg.GetMessage(), su); err != nil {
		o.deps.Log.Error("unmarshalling StyleUpdate", "error", err.Error())
		return
	}
	o.lock.Lock()
	matched := false
	for _, cfgSU := range o.cfg.Styles {
		if cfgSU.Selector == su.Selector && cfgSU.Property == su.Property {
			cfgSU.Value = su.Value
			matched = true
			break
		}
	}
	if !matched {
		o.cfg.Styles = append(o.cfg.Styles, su)
	}
	o.lock.Unlock()
	outMsg := &bus.BusMessage{
		Topic:   BusTopic_TRACKSTAR_OVERLAY_EVENT.String(),
		Type:    int32(MessageType_STYLE_UPDATE),
		Message: msg.Message,
	}
	o.deps.Bus.Send(outMsg)
}

func (o *Overlay) handleGetConfigRequest(msg *bus.BusMessage) {
	o.lock.Lock()
	b, err := proto.Marshal(&GetConfigResponse{
		Config: o.cfg,
	})
	o.lock.Unlock()
	if err != nil {
		o.deps.Log.Error("marshalling GetConfigResponse", "error", err.Error())
		return
	}
	reply := &bus.BusMessage{
		Type:    int32(MessageType_GET_CONFIG_RESPONSE),
		Message: b,
	}
	o.deps.Bus.SendReply(msg, reply)
}

package stagelinq

import (
	"context"

	"github.com/autonomouskoi/akcore/bus"
)

func (sl *StagelinQ) handleCommands(ctx context.Context) {
	sl.bus.HandleTypes(ctx, BusTopics_STAGELINQ_COMMAND.String(), 4,
		map[int32]bus.MessageHandler{
			int32(MessageTypeCommand_CONFIG_SET_REQ): sl.handleCommandConfigSet,
		},
		nil,
	)
}

func (sl *StagelinQ) handleCommandConfigSet(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.GetType() + 1,
	}
	csr := &ConfigSetRequest{}
	if reply.Error = sl.UnmarshalMessage(msg, csr); reply.Error != nil {
		return reply
	}
	sl.lock.Lock()
	sl.cfg = csr.Config
	sl.lock.Unlock()
	sl.MarshalMessage(reply, &ConfigSetResponse{Config: csr.Config})
	return reply
}

package stagelinq

import (
	"context"
	"slices"
	"strings"

	"github.com/autonomouskoi/akcore/bus"
)

func (sl *StagelinQ) handleRequests(ctx context.Context) {
	sl.bus.HandleTypes(ctx, BusTopics_STAGELINQ_REQUEST.String(), 4,
		map[int32]bus.MessageHandler{
			int32(MessageTypeRequest_CONFIG_GET_REQ):        sl.handleRequestGetConfig,
			int32(MessageTypeRequest_GET_DEVICES_REQ):       sl.handleRequestGetDevices,
			int32(MessageTypeRequest_CAPTURE_THRESHOLD_REQ): sl.handleRequestCaptureThreshold,
		},
		nil,
	)
}

func (sl *StagelinQ) handleRequestGetConfig(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.GetType() + 1,
	}
	sl.lock.Lock()
	sl.MarshalMessage(reply, &ConfigGetResponse{
		Config: sl.cfg,
	})
	sl.lock.Unlock()
	return reply
}

func (sl *StagelinQ) handleRequestGetDevices(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.GetType() + 1,
	}
	devices := []*Device{}
	sl.deviceStates.processDevices(func(m *Device) {
		devices = append(devices, m)
	})
	slices.SortFunc(devices, func(a, b *Device) int {
		return strings.Compare(a.GetToken(), b.GetToken())
	})
	sl.MarshalMessage(reply, &GetDevicesResponse{Devices: devices})
	return reply
}

func (sl *StagelinQ) handleRequestCaptureThreshold(msg *bus.BusMessage) *bus.BusMessage {
	reply := &bus.BusMessage{
		Topic: msg.GetTopic(),
		Type:  msg.GetType() + 1,
	}
	highest := 0.0
	for _, ds := range sl.deckStates {
		if ds.upfader > highest {
			highest = ds.upfader
		}
	}
	sl.cfg.FaderThreshold = highest
	sl.Log.Debug("setting threshold", "threshold", highest)
	sl.MarshalMessage(reply, &CaptureThresholdResponse{FaderThreshold: sl.cfg.FaderThreshold})
	return reply
}

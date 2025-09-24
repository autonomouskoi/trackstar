package stagelinq

import (
	"encoding/hex"
	"sync"

	"github.com/icedream/go-stagelinq"
	"google.golang.org/protobuf/proto"

	"github.com/autonomouskoi/akcore/bus"
	"github.com/autonomouskoi/akcore/svc/log"
)

type deviceStates struct {
	lock    sync.Mutex
	bus     *bus.Bus
	log     log.Logger
	devices map[string]*Device
}

func newDeviceStates(bus *bus.Bus, log log.Logger) *deviceStates {
	return &deviceStates{
		bus:     bus,
		log:     log,
		devices: map[string]*Device{},
	}
}

func slDeviceToken(slDevice *stagelinq.Device) string {
	token := slDevice.Token()
	return hex.EncodeToString(token[:])
}

func (d *deviceStates) update(token string, status DeviceStatus, detail string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.updateStateInternal(token, status, detail)
}

func (d *deviceStates) error(token, errStr string) {
	d.update(token, DeviceStatus_DEVICE_STATUS_ERROR, errStr)
}

func (d *deviceStates) updateStateInternal(token string, status DeviceStatus, detail string) {
	tsDevice := d.devices[token]
	if tsDevice == nil {
		d.log.Warn("updating state for missing device", "token", token,
			"status", status.String(),
			"detail", detail,
		)
		return
	}
	tsDevice.Status = status
	tsDevice.StatusDetail = detail

	msg := &bus.BusMessage{
		Topic: BusTopics_STAGELINQ_EVENT.String(),
		Type:  int32(MessageTypeEvent_DEVICE_STATE),
	}
	event := &DeviceStateEvent{
		Device: tsDevice,
	}
	var err error
	msg.Message, err = proto.Marshal(event)
	if err != nil {
		d.log.Error("marshalling DeviceStateEvent", "error", err.Error())
		return
	}
	d.log.Debug("sending device state",
		"token", token,
		"status", status.String(),
		"detail", detail,
	)
	d.bus.Send(msg)
}

// records the device as discovered, returning converted Device version and a
// bool indicating whether or not the device is new
func (d *deviceStates) discovered(slDevice *stagelinq.Device) (*Device, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()
	token := slDeviceToken(slDevice)
	tsDevice := d.devices[token]
	if tsDevice != nil {
		return tsDevice, false
	}

	tsDevice = &Device{
		Ip:              slDevice.IP.String(),
		Name:            slDevice.Name,
		SoftwareName:    slDevice.SoftwareName,
		SoftwareVerison: slDevice.SoftwareVersion,
		Token:           token,
	}
	d.devices[token] = tsDevice
	d.updateStateInternal(token, DeviceStatus_DEVICE_STATUS_UNKNOWN, "newly discovered")

	return tsDevice, true
}

func (d *deviceStates) setServices(token string, services []*stagelinq.Service) {
	d.lock.Lock()
	defer d.lock.Unlock()
	tsDevice := d.devices[token]
	if tsDevice == nil {
		d.log.Error("setting services on missing device", "token", token)
		return
	}
	for _, service := range services {
		tsDevice.Services = append(tsDevice.Services,
			&Service{
				Name: service.Name,
				Port: uint32(service.Port),
			},
		)
	}
}

func (d *deviceStates) processDevices(processor func(*Device)) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, device := range d.devices {
		processor(device)
	}
}

/*

type oldDiscovered struct {
	lock    sync.Mutex
	devices []*device
}

type device struct {
	sl *stagelinq.Device
	pb *Device
}

func (d *oldDiscovered) haveFound(slDevice *stagelinq.Device) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, foundDevice := range d.devices {
		if foundDevice.sl.IsEqual(slDevice) {
			return true
		}
	}
	d.devices = append(d.devices, &device{
		sl: slDevice,
		pb: convertSLDevice(slDevice),
	})
	return false
}

func (d *oldDiscovered) setServices(device *stagelinq.Device, services []*stagelinq.Service) {
	d.lock.Lock()
	defer d.lock.Unlock()
	var pb *Device
	for _, dev := range d.devices {
		if dev.sl.IsEqual(device) {
			pb = dev.pb
			break
		}
	}
	if pb == nil {
		return
	}
	for _, service := range services {
		pb.Services = append(pb.Services,
			&Service{
				Name: service.Name,
				Port: uint32(service.Port),
			},
		)
	}
}

func (d *oldDiscovered) processDevices(processor func([]*device)) {
	d.lock.Lock()
	defer d.lock.Unlock()
	processor(d.devices)
}
*/

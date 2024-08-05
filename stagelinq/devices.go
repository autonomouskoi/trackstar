package stagelinq

import (
	"sync"

	"github.com/icedream/go-stagelinq"
)

type discovered struct {
	lock    sync.Mutex
	devices []*device
}

type device struct {
	sl *stagelinq.Device
	pb *Device
}

func (d *discovered) haveFound(slDevice *stagelinq.Device) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, foundDevice := range d.devices {
		if foundDevice.sl.IsEqual(slDevice) {
			return true
		}
	}
	d.devices = append(d.devices, &device{
		sl: slDevice,
		pb: &Device{
			Ip:              slDevice.IP.String(),
			Name:            slDevice.Name,
			SoftwareName:    slDevice.SoftwareName,
			SoftwareVerison: slDevice.SoftwareVersion,
		},
	})
	return false
}

func (d *discovered) setServices(device *stagelinq.Device, services []*stagelinq.Service) {
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

func (d *discovered) processDevices(processor func([]*device)) {
	d.lock.Lock()
	defer d.lock.Unlock()
	processor(d.devices)
}

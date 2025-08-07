package prodjlink

import (
	"fmt"

	"go.evanpurkhiser.com/prolink"
)

type DeviceStatus struct {
	Status *prolink.CDJStatus
}

func DeckID(status *prolink.CDJStatus) string {
	return fmt.Sprintf("%d/%d/%d", status.PlayerID, status.TrackDevice, status.TrackSlot)
}

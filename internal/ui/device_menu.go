package ui

import "github.com/birgerrydback/wofi-bluetooth/internal/bluetooth"

// ShowDeviceMenu displays the menu for a specific device
func (ui UI) ShowDeviceMenu(device bluetooth.Device) {
	deviceOptions := ui.getDeviceMenuOptions(device)

	action := promptMenuOptions(deviceOptions, device.Name)

	switch action {
	case "":
		return
	case DeviceActionConnect:
		device.Connect()
		ui.ShowDeviceMenu(device)
	case DeviceActionDisconnect:
		device.Disconnect()
		ui.ShowDeviceMenu(device)
	case DeviceActionPair:
		device.Pair()
		ui.ShowDeviceMenu(device)
	case DeviceActionUnpair:
		device.Unpair()
		ui.ShowDeviceMenu(device)
	case DeviceActionTrust:
		device.SetTrust(true)
		ui.ShowDeviceMenu(device)
	case DeviceActionUntrust:
		device.SetTrust(false)
		ui.ShowDeviceMenu(device)
	case ActionGoBack:
		ui.ShowMainMenu()
	}
}

func (ui UI) getDeviceMenuOptions(device bluetooth.Device) []string {
	connectionAction := DeviceActionConnect
	if device.IsConnected() {
		connectionAction = DeviceActionDisconnect
	}

	pairingAction := DeviceActionPair
	if device.IsPaired() {
		pairingAction = DeviceActionUnpair
	}

	trustAction := DeviceActionTrust
	if device.IsTrusted() {
		trustAction = DeviceActionUntrust
	}

	return []string{connectionAction, pairingAction, trustAction, ActionGoBack}
}

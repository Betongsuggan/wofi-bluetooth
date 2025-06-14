package ui

import (
	"fmt"
)

// showMainMenu displays the main Bluetooth menu
func (ui UI) ShowMainMenu() {
	action := promptMenuOptions(ui.getMainMenuOptions(), "Bluetooth")

	switch action {
	case "":
		return
	case ActionDisableBluetooth:
		ui.bluetooth.SetPower(false)
		ui.ShowMainMenu()
	case ActionEnableBluetooth:
		ui.bluetooth.SetPower(true)
		ui.ShowMainMenu()
	case ActionScan:
		err := ui.bluetooth.SetScanning(true)
		if err != nil {
			fmt.Printf("failed to start scanning: %s", err)
			return
		}
		ui.ShowScanMenu()
	case ActionDisableDiscoverable:
		ui.bluetooth.SetDiscoverable(false)
		ui.ShowMainMenu()
	case ActionEnableDiscoverable:
		ui.bluetooth.SetDiscoverable(true)
		ui.ShowMainMenu()
	case ActionEnablePairable:
		ui.bluetooth.SetPairable(false)
		ui.ShowMainMenu()
	case ActionDisablePairable:
		ui.bluetooth.SetPairable(true)
		ui.ShowMainMenu()
	default:
		// Look if we matched any devices
		allDevices, _ := ui.bluetooth.GetKnownDevices()
		for _, device := range allDevices {
			if device.Name == action {
				ui.ShowDeviceMenu(device)
				return
			}
		}
	}
}

func (ui UI) getMainMenuOptions() []string {
	if !ui.bluetooth.IsPowered() {
		return []string{ActionEnableBluetooth}
	}

	devices, err := ui.bluetooth.GetKnownDevices()
	if err != nil {
		fmt.Println("Error getting all devices:", err)
	}

	deviceStrings := make([]string, 0)
	for _, device := range devices {
		deviceStrings = append(deviceStrings, device.Name)
	}

	pairableAction := ActionEnablePairable
	if ui.bluetooth.IsPairable() {
		pairableAction = ActionDisablePairable
	}

	discoverableAction := ActionEnableDiscoverable
	if ui.bluetooth.IsDiscoverable() {
		discoverableAction = ActionDisableDiscoverable
	}

	return append(deviceStrings, ActionScan, ActionDisableBluetooth, pairableAction, discoverableAction)
}

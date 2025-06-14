package ui

import (
	"fmt"
)

func (ui UI) ShowScanMenu() {
	action := promptMenuOptions(ui.getScanMenuOptions(), "Discovery")

	switch action {
	case "":
		ui.bluetooth.SetScanning(false)
		return
	case ActionGoBack:
		ui.bluetooth.SetScanning(false)
		ui.ShowMainMenu()
	case ActionRefresh:
		ui.ShowScanMenu()
	default:
		// Look if we matched any devices
		allDevices, _ := ui.bluetooth.GetUnknownDevices()
		for _, device := range allDevices {
			if device.Name == action {
				ui.ShowDeviceMenu(device)
				return
			}
		}
	}
}

func (ui UI) getScanMenuOptions() []string {
	options := []string{ActionRefresh}
	devices, err := ui.bluetooth.GetUnknownDevices()
	if err != nil {
		fmt.Println("Error getting all devices:", err)
	}

	for _, device := range devices {
		options = append(options, device.Name)
	}

	return append(options, ActionGoBack)
}

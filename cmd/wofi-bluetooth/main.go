// Package main provides the entry point for the wofi-bluetooth application
package main

import (
	"fmt"
	"strings"

	"github.com/birgerrydback/wofi-bluetooth/internal/bluetooth"
	"github.com/birgerrydback/wofi-bluetooth/internal/ui"
)

func main() {
	showMainMenu()
}

// showMainMenu displays the main Bluetooth menu
func showMainMenu() {
	controller := bluetooth.NewController()
	var options []string

	if controller.IsPowered() {
		power := "  Disable Bluetooth"

		// Get connected and available devices
		connectedDevices, err := controller.GetConnectedDevices()
		if err != nil {
			fmt.Println("Error getting connected devices:", err)
		}

		allDevices, err := controller.GetDevices()
		if err != nil {
			fmt.Println("Error getting all devices:", err)
		}

		// Create a map of connected device names for quick lookup
		connectedMap := make(map[string]bool)
		for _, device := range connectedDevices {
			connectedMap[device.Name] = true
			options = append(options, fmt.Sprintf("%s  %s", ui.ConnectedIcon, device.Name))
		}

		// Add disconnected devices
		for _, device := range allDevices {
			if !connectedMap[device.Name] && device.Name != "" {
				options = append(options, fmt.Sprintf("%s  %s", ui.DisconnectedIcon, device.Name))
			}
		}

		// Add divider and controller options
		options = append(options, power)

		// Add controller flags
		scanText := "Scan: off"
		if controller.IsScanning() {
			scanText = "Scan: on"
		}

		pairableText := "Pairable: off"
		if controller.IsPairable() {
			pairableText = "Pairable: on"
		}

		discoverableText := "Discoverable: off"
		if controller.IsDiscoverable() {
			discoverableText = "Discoverable: on"
		}

		options = append(options, scanText, pairableText, discoverableText, "Exit")
	} else {
		power := "󰂲  Enable Bluetooth"
		options = append(options, power, "Exit")
	}

	// Open wofi menu
	chosen := ui.ShowMenu(options, "Bluetooth")

	// Handle chosen option
	switch chosen {
	case "":
		fallthrough
	case "  Disable Bluetooth":
		controller.SetPower(false)
		showMainMenu()
	case "󰂲  Enable Bluetooth":
		controller.SetPower(true)
		showMainMenu()
	case "Scan: on":
		controller.SetScanning(false)
		showMainMenu()
	case "Scan: off":
		controller.SetScanning(true)
		showMainMenu()
	case "Discoverable: on":
		controller.SetDiscoverable(false)
		showMainMenu()
	case "Discoverable: off":
		controller.SetDiscoverable(true)
		showMainMenu()
	case "Pairable: on":
		controller.SetPairable(false)
		showMainMenu()
	case "Pairable: off":
		controller.SetPairable(true)
		showMainMenu()
	default:
		// Check if a device was selected
		cleanChosen := chosen
		if strings.HasPrefix(chosen, ui.ConnectedIcon) || strings.HasPrefix(chosen, ui.DisconnectedIcon) {
			cleanChosen = strings.TrimSpace(chosen[len(ui.ConnectedIcon):])
		}

		// Find the device in the list
		allDevices, _ := controller.GetDevices()
		for _, device := range allDevices {
			if device.Name == cleanChosen {
				showDeviceMenu(device)
				return
			}
		}
	}
}

// showDeviceMenu displays the menu for a specific device
func showDeviceMenu(device bluetooth.Device) {
	var options []string

	// Build connection status
	connectedText := "Connected: no"
	if device.IsConnected() {
		connectedText = "Connected: yes"
	}

	// Get paired and trusted status
	pairedText := "Paired: no"
	if device.IsPaired() {
		pairedText = "Paired: yes"
	}

	trustedText := "Trusted: no"
	if device.IsTrusted() {
		trustedText = "Trusted: yes"
	}

	// Build options
	options = append(options, connectedText, pairedText, trustedText, ui.GoBack, "Exit")

	// Open wofi menu
	chosen := ui.ShowMenu(options, device.Name)

	// Handle chosen option
	switch chosen {
	case "":
		fallthrough
	case "Connected: yes":
		device.Disconnect()
		showDeviceMenu(device)
	case "Connected: no":
		device.Connect()
		showDeviceMenu(device)
	case "Paired: yes":
		device.Unpair()
		showDeviceMenu(device)
	case "Paired: no":
		device.Pair()
		showDeviceMenu(device)
	case "Trusted: yes":
		device.SetTrust(false)
		showDeviceMenu(device)
	case "Trusted: no":
		device.SetTrust(true)
		showDeviceMenu(device)
	case ui.GoBack:
		showMainMenu()
	}
}

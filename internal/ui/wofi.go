// Package ui provides user interface functionality using wofi
package ui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/birgerrydback/wofi-bluetooth/internal/bluetooth"
)

const (
	// WofiCommand is the base command for launching wofi
	WofiCommand = "wofi -d -i -p"

	// ConnectedIcon is the icon for connected devices
	ConnectedIcon    = "󰂱"
	DisconnectedIcon = "󰾰"

	GoBack = "Back"
	Exit   = "Exit"
)

type (
	UI struct {
		bluetooth *bluetooth.Controller
	}
)

func NewUI(bluetooth *bluetooth.Controller) UI {
	return UI{
		bluetooth,
	}
}

// showMainMenu displays the main Bluetooth menu
func (ui UI) ShowMainMenu() {
	var options []string

	if ui.bluetooth.IsPowered() {
		power := "  Disable Bluetooth"

		// Get connected and available devices
		connectedDevices, err := ui.bluetooth.GetConnectedDevices()
		if err != nil {
			fmt.Println("Error getting connected devices:", err)
		}

		allDevices, err := ui.bluetooth.GetDevices()
		if err != nil {
			fmt.Println("Error getting all devices:", err)
		}

		// Create a map of connected device names for quick lookup
		connectedMap := make(map[string]bool)
		for _, device := range connectedDevices {
			connectedMap[device.Name] = true
			options = append(options, fmt.Sprintf("%s  %s", ConnectedIcon, device.Name))
		}

		// Add disconnected devices
		for _, device := range allDevices {
			if !connectedMap[device.Name] && device.Name != "" {
				options = append(options, fmt.Sprintf("%s  %s", DisconnectedIcon, device.Name))
			}
		}

		// Add divider and controller options
		options = append(options, power)

		// Add controller flags
		scanText := "Scan: off"
		if ui.bluetooth.IsScanning() {
			scanText = "Scan: on"
		}

		pairableText := "Pairable: off"
		if ui.bluetooth.IsPairable() {
			pairableText = "Pairable: on"
		}

		discoverableText := "Discoverable: off"
		if ui.bluetooth.IsDiscoverable() {
			discoverableText = "Discoverable: on"
		}

		options = append(options, scanText, pairableText, discoverableText, "Exit")
	} else {
		power := "󰂲  Enable Bluetooth"
		options = append(options, power, "Exit")
	}

	// Open wofi menu
	chosen := promptMenuOptions(options, "Bluetooth")

	// Handle chosen option
	switch chosen {
	case "":
		fallthrough
	case "  Disable Bluetooth":
		ui.bluetooth.SetPower(false)
		ui.ShowMainMenu()
	case "󰂲  Enable Bluetooth":
		ui.bluetooth.SetPower(true)
		ui.ShowMainMenu()
	case "Scan: on":
		ui.bluetooth.SetScanning(false)
		ui.ShowMainMenu()
	case "Scan: off":
		ui.bluetooth.SetScanning(true)
		ui.ShowMainMenu()
	case "Discoverable: on":
		ui.bluetooth.SetDiscoverable(false)
		ui.ShowMainMenu()
	case "Discoverable: off":
		ui.bluetooth.SetDiscoverable(true)
		ui.ShowMainMenu()
	case "Pairable: on":
		ui.bluetooth.SetPairable(false)
		ui.ShowMainMenu()
	case "Pairable: off":
		ui.bluetooth.SetPairable(true)
		ui.ShowMainMenu()
	default:
		// Check if a device was selected
		cleanChosen := chosen
		if strings.HasPrefix(chosen, ConnectedIcon) || strings.HasPrefix(chosen, DisconnectedIcon) {
			cleanChosen = strings.TrimSpace(chosen[len(ConnectedIcon):])
		}

		// Find the device in the list
		allDevices, _ := ui.bluetooth.GetDevices()
		for _, device := range allDevices {
			if device.Name == cleanChosen {
				ui.showDeviceMenu(device)
				return
			}
		}
	}
}

// showDeviceMenu displays the menu for a specific device
func (ui UI) showDeviceMenu(device bluetooth.Device) {
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
	options = append(options, connectedText, pairedText, trustedText, GoBack, "Exit")

	// Open wofi menu
	chosen := promptMenuOptions(options, device.Name)

	// Handle chosen option
	switch chosen {
	case "":
		fallthrough
	case "Connected: yes":
		device.Disconnect()
		ui.showDeviceMenu(device)
	case "Connected: no":
		device.Connect()
		ui.showDeviceMenu(device)
	case "Paired: yes":
		device.Unpair()
		ui.showDeviceMenu(device)
	case "Paired: no":
		device.Pair()
		ui.showDeviceMenu(device)
	case "Trusted: yes":
		device.SetTrust(false)
		ui.showDeviceMenu(device)
	case "Trusted: no":
		device.SetTrust(true)
		ui.showDeviceMenu(device)
	case GoBack:
		ui.ShowMainMenu()
	}
}

// promptMenuOptions displays a wofi menu with the given options and returns the chosen option
func promptMenuOptions(options []string, prompt string) string {
	// Create a temporary file for the menu options
	tmpfile, err := os.CreateTemp("", "wofi-bluetooth-menu-")
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return ""
	}
	defer os.Remove(tmpfile.Name())

	// Write options to the temporary file
	for _, option := range options {
		fmt.Fprintln(tmpfile, option)
	}
	tmpfile.Close()

	// Count lines for wofi height
	lines := len(options)

	// Run wofi command
	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s \"%s\" -L %d < %s", WofiCommand, prompt, lines, tmpfile.Name()))
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		// User probably cancelled
		return ""
	}

	// Return the chosen option (trimming newline)
	return strings.TrimSpace(out.String())
}

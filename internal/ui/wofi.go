// Package ui provides user interface functionality using wofi
package ui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/birgerrydback/wofi-bluetooth/internal/bluetooth"
)

const (
	// WofiCommand is the base command for launching wofi
	WofiCommand = "wofi -d -i -p"

	// Icons for different device states
	ConnectedIcon    = "󰂱"
	DisconnectedIcon = "󰾰"
	DiscoveredIcon   = "󰑐"

	ActionDisableBluetooth    = "  Disable Bluetooth"
	ActionEnableBluetooth     = "󰂲  Enable Bluetooth"
	ActionEnableDiscoverable  = "  Enable discoverable"
	ActionDisableDiscoverable = "  Disable discoverable"
	ActionEnablePairable      = "󰌺  Enable pairable"
	ActionDisablePairable     = "  Disable pairable"
	ActionScan                = "󱉶  Scan"
	ActionGoBack              = "Back"
	ActionExit                = "Exit"

	DeviceActionPair       = "󰌺  Pair"
	DeviceActionUnpair     = "  Unpair"
	DeviceActionTrust      = "󱚩  Pair"
	DeviceActionUntrust    = "󱎚  Unpair"
	DeviceActionConnect    = "󰂲  Connect"
	DeviceActionDisconnect = "󰂱  Disconnect"
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
		power := ActionDisableBluetooth

		// Get connected and available devices
		connectedDevices, err := ui.bluetooth.GetConnectedDevices()
		if err != nil {
			fmt.Println("Error getting connected devices:", err)
		}

		allDevices, err := ui.bluetooth.GetDevices()
		if err != nil {
			fmt.Println("Error getting all devices:", err)
		}

		// Get discovered but not paired devices if scanning is on
		var discoveredDevices []bluetooth.Device
		if ui.bluetooth.IsScanning() {
			discoveredDevices, err = ui.bluetooth.GetDiscoveredDevices()
			if err != nil {
				fmt.Println("Error getting discovered devices:", err)
			}
		}

		// Create maps for quick lookup
		connectedMap := make(map[string]bool)
		knownDevicesMap := make(map[string]bool)

		// Add connected devices to the menu
		for _, device := range connectedDevices {
			connectedMap[device.Name] = true
			knownDevicesMap[device.MAC] = true
			options = append(options, fmt.Sprintf("%s  %s", ConnectedIcon, device.Name))
		}

		// Add disconnected but paired devices
		for _, device := range allDevices {
			if !connectedMap[device.Name] && device.Name != "" {
				knownDevicesMap[device.MAC] = true
				options = append(options, fmt.Sprintf("%s  %s", DisconnectedIcon, device.Name))
			}
		}

		// Add discovered but not paired devices if scanning is on
		for _, device := range discoveredDevices {
			if !knownDevicesMap[device.MAC] && device.Name != "" {
				options = append(options, fmt.Sprintf("%s  %s", DiscoveredIcon, device.Name))
			}
		}

		// Add divider and controller options
		options = append(options, power)

		// Add controller flags
		scanAction := ActionScan

		pairableAction := ActionEnablePairable
		if ui.bluetooth.IsPairable() {
			pairableAction = ActionDisablePairable
		}

		discoverableAction := ActionEnableDiscoverable
		if ui.bluetooth.IsDiscoverable() {
			discoverableAction = ActionDisableDiscoverable
		}

		options = append(options, scanAction, pairableAction, discoverableAction, ActionExit)
	} else {
		power := ActionEnableBluetooth
		options = append(options, power, ActionExit)
	}

	// Open wofi menu
	action := promptMenuOptions(options, "Bluetooth")

	// Handle chosen option
	switch action {
	case "":
		// User pressed Escape, just exit
		return
	case ActionDisableBluetooth:
		ui.bluetooth.SetPower(false)
		ui.ShowMainMenu()
	case ActionEnableBluetooth:
		ui.bluetooth.SetPower(true)
		ui.ShowMainMenu()
	case ActionScan:
		ui.bluetooth.SetScanning(true)

		// Wait a moment to let the scan start
		time.Sleep(500 * time.Millisecond)

		ui.ShowMainMenu()
	// case "Scan: off":
	//	// Start scanning with the custom command sequence
	//	ui.bluetooth.SetScanning(true)

	//	// Refresh the menu to show discovered devices
	//	ui.ShowMainMenu()
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
		// Check if a device was selected
		cleanChosen := action
		var isDiscovered bool

		if strings.HasPrefix(action, ConnectedIcon) || strings.HasPrefix(action, DisconnectedIcon) {
			cleanChosen = strings.TrimSpace(action[len(ConnectedIcon):])
		} else if strings.HasPrefix(action, DiscoveredIcon) {
			cleanChosen = strings.TrimSpace(action[len(DiscoveredIcon):])
			isDiscovered = true
		}

		// Find the device in the appropriate list
		if isDiscovered {
			// Look in discovered devices
			discoveredDevices, _ := ui.bluetooth.GetDiscoveredDevices()
			for _, device := range discoveredDevices {
				if device.Name == cleanChosen {
					ui.showDiscoveredDeviceMenu(device)
					return
				}
			}
		} else {
			// Look in paired devices
			allDevices, _ := ui.bluetooth.GetDevices()
			for _, device := range allDevices {
				if device.Name == cleanChosen {
					ui.showDeviceMenu(device)
					return
				}
			}
		}
	}
}

// showDeviceMenu displays the menu for a specific device
func (ui UI) showDeviceMenu(device bluetooth.Device) {
	var options []string

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

	options = append(options, connectionAction, pairingAction, trustAction, ActionGoBack, ActionExit)

	chosen := promptMenuOptions(options, device.Name)

	// Handle chosen option
	switch chosen {
	case "":
		fallthrough
	case DeviceActionConnect:
		device.Connect()
		ui.showDeviceMenu(device)
	case DeviceActionDisconnect:
		device.Disconnect()
		ui.showDeviceMenu(device)
	case DeviceActionPair:
		device.Pair()
		ui.showDeviceMenu(device)
	case DeviceActionUnpair:
		device.Unpair()
		ui.showDeviceMenu(device)
	case DeviceActionTrust:
		device.SetTrust(true)
		ui.showDeviceMenu(device)
	case DeviceActionUntrust:
		device.SetTrust(false)
		ui.showDeviceMenu(device)
	case ActionGoBack:
		ui.ShowMainMenu()
	}
}

// showDiscoveredDeviceMenu displays the menu for a discovered device
func (ui UI) showDiscoveredDeviceMenu(device bluetooth.Device) {
	var options []string

	// Build options for discovered devices
	options = append(options, "Pair", "Pair and Trust", ActionGoBack, "Exit")

	// Open wofi menu
	chosen := promptMenuOptions(options, device.Name)

	// Handle chosen option
	switch chosen {
	case "":
		ui.ShowMainMenu()
	case "Pair":
		device.Pair()
		ui.ShowMainMenu()
	case "Pair and Trust":
		device.Pair()
		device.SetTrust(true)
		ui.ShowMainMenu()
	case ActionGoBack:
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

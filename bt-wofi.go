package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	divider          = "---------"
	goBack           = "Back"
	wofiCommand      = "wofi -d -i -p"
	connectedIcon    = "󰂱"
	disconnectedIcon = "󰾰"
)

// Execute a command and return its output as a string
func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

// Check if bluetooth controller is powered on
func isPowerOn() bool {
	output, err := execCommand("bluetoothctl", "show")
	if err != nil {
		fmt.Println("Error checking power state:", err)
		return false
	}
	return strings.Contains(output, "Powered: yes")
}

// Check if controller is scanning for new devices
func isScanOn() (string, bool) {
	output, err := execCommand("bluetoothctl", "show")
	if err != nil {
		fmt.Println("Error checking scan state:", err)
		return "Scan: off", false
	}
	if strings.Contains(output, "Discovering: yes") {
		return "Scan: on", true
	}
	return "Scan: off", false
}

// Toggle scanning state
func toggleScan() {
	// scanText, isScanning := isScanOn()
	//if isScanning {
	//	// Kill any running scan processes
	//	exec.Command("pkill", "-f", "bluetoothctl scan on").Run()
	//	execCommand("bluetoothctl", "scan", "off")
	//	showMenu()
	//} else {
	//	// Start scanning in background
	//	cmd := exec.Command("bluetoothctl", "scan", "on")
	//	cmd.Start() // Don't wait for it to complete
	//	fmt.Println("Scanning...")
	//	// Sleep for 5 seconds to allow some devices to be discovered
	//	execCommand("sleep", "5")
	//	showMenu()
	//}
}

// Check if controller is pairable
func isPairableOn() (string, bool) {
	output, err := execCommand("bluetoothctl", "show")
	if err != nil {
		fmt.Println("Error checking pairable state:", err)
		return "Pairable: off", false
	}
	if strings.Contains(output, "Pairable: yes") {
		return "Pairable: on", true
	}
	return "Pairable: off", false
}

// Toggle pairable state
func togglePairable() {
	_, isPairable := isPairableOn()
	if isPairable {
		execCommand("bluetoothctl", "pairable", "off")
	} else {
		execCommand("bluetoothctl", "pairable", "on")
	}
	showMenu()
}

// Check if controller is discoverable
func isDiscoverableOn() (string, bool) {
	output, err := execCommand("bluetoothctl", "show")
	if err != nil {
		fmt.Println("Error checking discoverable state:", err)
		return "Discoverable: off", false
	}
	if strings.Contains(output, "Discoverable: yes") {
		return "Discoverable: on", true
	}
	return "Discoverable: off", false
}

// Toggle discoverable state
func toggleDiscoverable() {
	_, isDiscoverable := isDiscoverableOn()
	if isDiscoverable {
		execCommand("bluetoothctl", "discoverable", "off")
	} else {
		execCommand("bluetoothctl", "discoverable", "on")
	}
	showMenu()
}

// Check if a device is connected
func isDeviceConnected(mac string) bool {
	output, err := execCommand("bluetoothctl", "info", mac)
	if err != nil {
		fmt.Println("Error checking device connection:", err)
		return false
	}
	return strings.Contains(output, "Connected: yes")
}

// Toggle device connection
func toggleConnection(mac, device string) {
	if isDeviceConnected(mac) {
		execCommand("bluetoothctl", "disconnect", mac)
	} else {
		execCommand("bluetoothctl", "connect", mac)
	}
	deviceMenu(device)
}

// Check if a device is paired
func isDevicePaired(mac string) (string, bool) {
	output, err := execCommand("bluetoothctl", "info", mac)
	if err != nil {
		fmt.Println("Error checking device pairing:", err)
		return "Paired: no", false
	}
	if strings.Contains(output, "Paired: yes") {
		return "Paired: yes", true
	}
	return "Paired: no", false
}

// Toggle device paired state
func togglePaired(mac, device string) {
	_, isPaired := isDevicePaired(mac)
	if isPaired {
		execCommand("bluetoothctl", "remove", mac)
	} else {
		execCommand("bluetoothctl", "pair", mac)
	}
	deviceMenu(device)
}

// Check if a device is trusted
func isDeviceTrusted(mac string) (string, bool) {
	output, err := execCommand("bluetoothctl", "info", mac)
	if err != nil {
		fmt.Println("Error checking device trust:", err)
		return "Trusted: no", false
	}
	if strings.Contains(output, "Trusted: yes") {
		return "Trusted: yes", true
	}
	return "Trusted: no", false
}

// Toggle device trust state
func toggleTrust(mac, device string) {
	_, isTrusted := isDeviceTrusted(mac)
	if isTrusted {
		execCommand("bluetoothctl", "untrust", mac)
	} else {
		execCommand("bluetoothctl", "trust", mac)
	}
	deviceMenu(device)
}

// Toggle power state
func togglePower() {
	if isPowerOn() {
		execCommand("bluetoothctl", "power", "off")
	} else {
		// Check if bluetooth is blocked by rfkill
		output, _ := execCommand("rfkill", "list", "bluetooth")
		if strings.Contains(output, "blocked: yes") {
			execCommand("rfkill", "unblock", "bluetooth")
			execCommand("sleep", "3") // Wait for bluetooth to initialize
		}
		execCommand("bluetoothctl", "power", "on")
	}
	showMenu()
}

// Show a submenu for a specific device
func deviceMenu(device string) {
	parts := strings.SplitN(device, " ", 3)
	if len(parts) < 3 {
		fmt.Println("Invalid device format:", device)
		return
	}

	mac := parts[1]
	deviceName := parts[2]

	var options []string

	// Build connection status
	connected := "Connected: no"
	if isDeviceConnected(mac) {
		connected = "Connected: yes"
	}

	// Get paired and trusted status
	paired, _ := isDevicePaired(mac)
	trusted, _ := isDeviceTrusted(mac)

	// Build options
	options = append(options, connected, paired, trusted, divider, goBack, "Exit")

	// Open wofi menu
	chosen := showWofiMenu(options, deviceName)

	// Handle chosen option
	switch chosen {
	case "":
		fallthrough
	case divider:
		fmt.Println("No option chosen.")
	case connected:
		toggleConnection(mac, device)
	case paired:
		togglePaired(mac, device)
	case trusted:
		toggleTrust(mac, device)
	case goBack:
		showMenu()
	}
}

// Show the main wofi menu
func showMenu() {
	var options []string

	if isPowerOn() {
		power := "  Disable Bluetooth"

		// Get connected and available devices
		connectedOutput, _ := execCommand("bluetoothctl", "devices", "Connected")
		devicesOutput, _ := execCommand("bluetoothctl", "devices")

		// Process connected devices
		connectedDevices := make(map[string]bool)
		scanner := bufio.NewScanner(strings.NewReader(connectedOutput))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Device ") {
				deviceName := strings.Join(strings.Split(line, " ")[2:], " ")
				if deviceName != "" {
					connectedDevices[deviceName] = true
					options = append(options, fmt.Sprintf("%s  %s", connectedIcon, deviceName))
				}
			}
		}

		// Process all devices and filter out connected ones
		scanner = bufio.NewScanner(strings.NewReader(devicesOutput))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Device ") {
				deviceName := strings.Join(strings.Split(line, " ")[2:], " ")
				if deviceName != "" && !connectedDevices[deviceName] {
					options = append(options, fmt.Sprintf("%s  %s", disconnectedIcon, deviceName))
				}
			}
		}

		// Add divider and controller options
		options = append(options, divider, power)

		// Get controller flags
		scan, _ := isScanOn()
		pairable, _ := isPairableOn()
		discoverable, _ := isDiscoverableOn()

		options = append(options, scan, pairable, discoverable, "Exit")
	} else {
		power := "󰂲  Enable Bluetooth"
		options = append(options, power, "Exit")
	}

	// Open wofi menu
	chosen := showWofiMenu(options, "Bluetooth")

	// Handle chosen option
	switch chosen {
	case "":
		fallthrough
	case divider:
		fmt.Println("No option chosen.")
	case " Disable Bluetooth", "󰂲  Enable Bluetooth":
		togglePower()
	case "Scan: on", "Scan: off":
		toggleScan()
	case "Discoverable: on", "Discoverable: off":
		toggleDiscoverable()
	case "Pairable: on", "Pairable: off":
		togglePairable()
	default:
		// Check if a device was selected
		deviceLine := ""

		// Strip the icon prefix if present
		cleanChosen := chosen
		if strings.HasPrefix(chosen, connectedIcon) || strings.HasPrefix(chosen, disconnectedIcon) {
			cleanChosen = strings.TrimSpace(chosen[len(connectedIcon):])
		}

		// Find the device in the list
		devicesOutput, _ := execCommand("bluetoothctl", "devices")
		scanner := bufio.NewScanner(strings.NewReader(devicesOutput))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, cleanChosen) {
				deviceLine = line
				break
			}
		}

		if deviceLine != "" {
			deviceMenu(deviceLine)
		}
	}
}

// Show a wofi menu with the given options and return the chosen option
func showWofiMenu(options []string, prompt string) string {
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
	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s \"%s\" -L %d < %s", wofiCommand, prompt, lines, tmpfile.Name()))
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

func main() {
	showMenu()
}

// Package bluetooth provides functionality for interacting with the Bluetooth system
package bluetooth

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const (
	DeviceStatusConnected DeviceStatus = iota
	DeviceStatusPaired
	DeviceStatusDiscovered
)

type (
	Controller struct{}

	// Device represents a Bluetooth device
	Device struct {
		MAC    string
		Name   string
		Line   string // Original line from bluetoothctl
		Status DeviceStatus
	}

	DeviceStatus int
)

func NewController() *Controller {
	return &Controller{}
}

// IsPowered checks if bluetooth controller is powered on
func (c *Controller) IsPowered() bool {
	output, err := execCommand("bluetoothctl", "show")
	if err != nil {
		fmt.Println("Error checking power state:", err)
		return false
	}
	return strings.Contains(output, "Powered: yes")
}

// SetPower sets the power state of the Bluetooth controller
func (c *Controller) SetPower(on bool) error {
	state := "off"
	if on {
		state = "on"
		output, _ := execCommand("rfkill", "list", "bluetooth")
		if strings.Contains(output, "blocked: yes") {
			execCommand("rfkill", "unblock", "bluetooth")
			execCommand("sleep", "3") // Wait for bluetooth to initialize
		}
	}
	_, err := execCommand("bluetoothctl", "power", state)
	return err
}

// IsScanning checks if controller is scanning for new devices
func (c *Controller) IsScanning() bool {
	// Check if bluetoothctl is in scanning mode
	output, err := execCommand("bluetoothctl", "show")
	if err != nil {
		fmt.Println("Error checking scan state:", err)
		return false
	}

	// Also check if our scanning process is running
	scanningProcess, _ := execCommand("pgrep", "-f", "echo -e 'power on\\nscan on\\n' | bluetoothctl")

	return strings.Contains(output, "Discovering: yes") || scanningProcess != ""
}

// SetScanning sets the scanning state
func (c *Controller) SetScanning(on bool) error {
	if on {
		// Execute the specific scanning sequence
		cmd := exec.Command("bash", "-c", `
			echo -e 'power on\nscan on\n' | bluetoothctl
			sleep 5
			echo -e 'scan off\ndevices\nquit' | bluetoothctl
		`)

		// Run in background
		err := cmd.Start()
		if err != nil {
			return err
		}

		// No need to wait for completion as we want to return to the UI immediately
		// The scan will complete on its own after 5 seconds
	} else {
		// Kill any running scan processes
		exec.Command("pkill", "-f", "bluetoothctl scan on").Run()
		_, err := execCommand("bluetoothctl", "scan", "off")
		if err != nil {
			return err
		}
	}
	return nil
}

// IsPairable checks if controller is pairable
func (c *Controller) IsPairable() bool {
	output, err := execCommand("bluetoothctl", "show")
	if err != nil {
		fmt.Println("Error checking pairable state:", err)
		return false
	}
	return strings.Contains(output, "Pairable: yes")
}

// SetPairable sets the pairable state
func (c *Controller) SetPairable(on bool) error {
	state := "off"
	if on {
		state = "on"
	}
	_, err := execCommand("bluetoothctl", "pairable", state)
	return err
}

// IsDiscoverable checks if controller is discoverable
func (c *Controller) IsDiscoverable() bool {
	output, err := execCommand("bluetoothctl", "show")
	if err != nil {
		fmt.Println("Error checking discoverable state:", err)
		return false
	}
	return strings.Contains(output, "Discoverable: yes")
}

// SetDiscoverable sets the discoverable state
func (c *Controller) SetDiscoverable(on bool) error {
	state := "off"
	if on {
		state = "on"
	}
	_, err := execCommand("bluetoothctl", "discoverable", state)
	return err
}

// IsConnected checks if a device is connected
func (d *Device) IsConnected() bool {
	output, err := execCommand("bluetoothctl", "info", d.MAC)
	if err != nil {
		fmt.Println("Error checking device connection:", err)
		return false
	}
	return strings.Contains(output, "Connected: yes")
}

// Connect connects to a device
func (d *Device) Connect() error {
	_, err := execCommand("bluetoothctl", "connect", d.MAC)
	return err
}

// Disconnect disconnects from a device
func (d *Device) Disconnect() error {
	_, err := execCommand("bluetoothctl", "disconnect", d.MAC)
	return err
}

// IsPaired checks if a device is paired
func (d *Device) IsPaired() bool {
	output, err := execCommand("bluetoothctl", "info", d.MAC)
	if err != nil {
		fmt.Println("Error checking device pairing:", err)
		return false
	}
	return strings.Contains(output, "Paired: yes")
}

// Pair pairs with a device
func (d *Device) Pair() error {
	_, err := execCommand("bluetoothctl", "pair", d.MAC)
	return err
}

// Unpair unpairs from a device
func (d *Device) Unpair() error {
	_, err := execCommand("bluetoothctl", "remove", d.MAC)
	return err
}

// IsTrusted checks if a device is trusted
func (d *Device) IsTrusted() bool {
	output, err := execCommand("bluetoothctl", "info", d.MAC)
	if err != nil {
		fmt.Println("Error checking device trust:", err)
		return false
	}
	return strings.Contains(output, "Trusted: yes")
}

// SetTrust sets the trust state of a device
func (d *Device) SetTrust(trusted bool) error {
	cmd := "untrust"
	if trusted {
		cmd = "trust"
	}
	_, err := execCommand("bluetoothctl", cmd, d.MAC)
	return err
}

// GetDevices returns a list of all known devices
func (c *Controller) GetDevices() ([]Device, error) {
	output, err := execCommand("bluetoothctl", "devices")
	if err != nil {
		return nil, err
	}

	var devices []Device
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Device ") {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 3 {
				devices = append(devices, Device{
					MAC:  parts[1],
					Name: parts[2],
					Line: line,
				})
			}
		}
	}

	return devices, nil
}

// GetConnectedDevices returns a list of connected devices
func (c *Controller) GetConnectedDevices() ([]Device, error) {
	output, err := execCommand("bluetoothctl", "devices", "Connected")
	if err != nil {
		return nil, err
	}

	var devices []Device
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Device ") {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 3 {
				devices = append(devices, Device{
					MAC:    parts[1],
					Name:   parts[2],
					Line:   line,
					Status: DeviceStatusConnected,
				})
			}
		}
	}

	return devices, nil
}

// GetDiscoveredDevices returns a list of discovered but not paired devices
func (c *Controller) GetDiscoveredDevices() ([]Device, error) {
	// First, get all devices from bluetoothctl
	allDevicesOutput, err := execCommand("bluetoothctl", "devices")
	if err != nil {
		return nil, err
	}

	// Then get paired devices
	pairedDevicesOutput, err := execCommand("bluetoothctl", "paired-devices")
	if err != nil {
		return nil, err
	}

	// Create a map of paired device MACs for quick lookup
	pairedMACs := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(pairedDevicesOutput))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Device ") {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 2 {
				pairedMACs[parts[1]] = true
			}
		}
	}

	// Filter for discovered but not paired devices
	var discoveredDevices []Device
	scanner = bufio.NewScanner(strings.NewReader(allDevicesOutput))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Device ") {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 3 {
				mac := parts[1]
				if !pairedMACs[mac] {
					discoveredDevices = append(discoveredDevices, Device{
						MAC:    mac,
						Name:   parts[2],
						Line:   line,
						Status: DeviceStatusDiscovered,
					})
				}
			}
		}
	}

	return discoveredDevices, nil
}

func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

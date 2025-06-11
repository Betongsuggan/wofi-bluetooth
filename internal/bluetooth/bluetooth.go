// Package bluetooth provides functionality for interacting with the Bluetooth system
package bluetooth

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Controller represents a Bluetooth controller
type Controller struct{}

// NewController creates a new Bluetooth controller instance
func NewController() *Controller {
	return &Controller{}
}

// Execute a command and return its output as a string
func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
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
		// Check if bluetooth is blocked by rfkill
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
	output, err := execCommand("bluetoothctl", "show")
	if err != nil {
		fmt.Println("Error checking scan state:", err)
		return false
	}
	return strings.Contains(output, "Discovering: yes")
}

// SetScanning sets the scanning state
func (c *Controller) SetScanning(on bool) error {
	if on {
		// Start scanning in background
		cmd := exec.Command("bluetoothctl", "scan", "on")
		err := cmd.Start() // Don't wait for it to complete
		if err != nil {
			return err
		}
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

// Device represents a Bluetooth device
type Device struct {
	MAC  string
	Name string
	Line string // Original line from bluetoothctl
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
					MAC:  parts[1],
					Name: parts[2],
					Line: line,
				})
			}
		}
	}
	
	return devices, nil
}

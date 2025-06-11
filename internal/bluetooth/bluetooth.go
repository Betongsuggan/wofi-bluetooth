package bluetooth

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

const (
	StateOn  = "on"
	StateOff = "off"
)

type (
	Controller struct{}
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
	state := StateOff
	if on {
		state = StateOn
		output, _ := execCommand("rfkill", "list", "bluetooth")
		if strings.Contains(output, "blocked: yes") {
			execCommand("rfkill", "unblock", "bluetooth")
			execCommand("sleep", "3") // Wait for bluetooth to initialize
		}
	}
	_, err := execCommand("bluetoothctl", "power", state)
	return err
}

// SetScanning sets the scanning state
func (c *Controller) SetScanning(durationSeconds int) error {
	// Kill any running scan processes
	exec.Command("pkill", "-f", "bluetoothctl scan on").Run()
	_, err := execCommand("bluetoothctl", "scan", "off")
	if err != nil {
		return err
	}

	// Execute the specific scanning sequence
	cmd := exec.Command("bash", "-c", `
			echo -e 'power on\nscan on\n' | bluetoothctl
			sleep 5
			echo -e 'scan off\ndevices\nquit' | bluetoothctl
		`)

	// Run in background
	err = cmd.Start()
	if err != nil {
		return err
	}

	// No need to wait for completion as we want to return to the UI immediately
	// The scan will complete on its own after 5 seconds

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
	state := StateOff
	if on {
		state = StateOn
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
	state := StateOff
	if on {
		state = StateOn
	}
	_, err := execCommand("bluetoothctl", "discoverable", state)
	return err
}

// GetDevices returns a list of all known devices
func (c *Controller) GetDevices() ([]Device, error) {
	connectedDevices, err := execCommand("bluetoothctl", "devices", "Connected")
	if err != nil {
		return nil, err
	}

	allDevices, err := execCommand("bluetoothctl", "devices")
	if err != nil {
		return nil, err
	}

	devices := scanDevices(connectedDevices, DeviceStatusConnected)
	paired := scanDevices(allDevices, DeviceStatusPaired)

	for _, device := range paired {
		if slices.ContainsFunc(devices, func(other Device) bool { return device.Equals(other) }) {
			continue
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func scanDevices(output string, status DeviceStatus) []Device {
	var devices []Device
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Device ") {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 3 {
				devices = append(devices, NewDevice(
					parts[2],
					parts[1],
					line,
					status,
					DeviceTypeLaptop,
				))
			}
		}
	}
	return devices
}

func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

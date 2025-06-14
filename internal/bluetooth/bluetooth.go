package bluetooth

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
	"time"
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

// Scan sets the scanning state
func (c *Controller) Scan(duration time.Duration, devices chan Device) error {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	// Kill any running scan processes
	//exec.Command("pkill", "-f", "bluetoothctl scan on").Run()
	//_, err := execCommand()
	//_, err := execCommand("bluetoothctl", "scan", "off")
	//if err != nil {
	//	return fmt.Errorf("failed to kill existing scanning processes: %w", err)
	//}

	// Execute the specific scanning sequence
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf(`
			echo -e 'power on\nscan on\n' | bluetoothctl
			sleep %f
			echo -e 'scan off\ndevices\nquit' | bluetoothctl
		`, duration.Seconds()))

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	// Run in background
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to execute scan command: %w", err)
	}
	go func() {
		io.WriteString(stdin, "scan on\n")
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[NEW] Device") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				mac := parts[3]
				name := strings.Join(parts[4:], " ")
				devices <- NewDevice(name, mac, line, DeviceStatusDiscovered, DeviceTypeGeneric)
			}
		}
	}

	return cmd.Wait()
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
		cmd := exec.Command("bluetoothctl", "scan", "on")
		err := cmd.Start()
		if err != nil {
			return err
		}
	} else {
		exec.Command("pkill", "-f", "bluetoothctl scan on").Run()
		_, err := execCommand("bluetoothctl", "scan", "off")
		if err != nil {
			return err
		}
	}
	return nil
}

// GetKnownDevices returns a list of all known devices
func (c *Controller) GetKnownDevices() ([]Device, error) {
	connectedDevices, err := execCommand("bluetoothctl", "devices", "Connected")
	if err != nil {
		return nil, err
	}

	pairedDevices, err := execCommand("bluetoothctl", "devices", "Paired")
	if err != nil {
		return nil, err
	}

	trustedDevices, err := execCommand("bluetoothctl", "devices", "Trusted")
	if err != nil {
		return nil, err
	}

	devices := parseDevices(connectedDevices, DeviceStatusConnected)
	paired := parseDevices(pairedDevices, DeviceStatusPaired)
	trusted := parseDevices(trustedDevices, DeviceStatusTrusted)

	for _, device := range paired {
		if slices.ContainsFunc(devices, func(other Device) bool { return device.Equals(other) }) {
			continue
		}
		devices = append(devices, device)
	}

	for _, device := range trusted {
		if slices.ContainsFunc(devices, func(other Device) bool { return device.Equals(other) }) {
			continue
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (c *Controller) GetUnknownDevices() ([]Device, error) {
	knownDevices, err := c.GetKnownDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get known devices: %w", err)
	}

	allDeviceStrings, err := execCommand("bluetoothctl", "devices")
	if err != nil {
		return nil, fmt.Errorf("failed to list all devices: %w", err)
	}
	allDevices := parseDevices(allDeviceStrings, DeviceStatusConnected)

	unknownDevices := make([]Device, 0)
	for _, device := range allDevices {
		if slices.ContainsFunc(knownDevices, func(other Device) bool { return device.Equals(other) }) {
			continue
		}
		unknownDevices = append(unknownDevices, device)
	}

	return unknownDevices, nil
}

func parseDevices(output string, status DeviceStatus) []Device {
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
	if err != nil {
		return "", fmt.Errorf("command %s with arguments %s failed to run: %w", name, args, err)
	}
	return out.String(), nil
}

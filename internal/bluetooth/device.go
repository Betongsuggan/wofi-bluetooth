package bluetooth

import (
	"fmt"
	"strings"
)

const (
	DeviceStatusConnected DeviceStatus = iota
	DeviceStatusPaired
	DeviceStatusTrusted
	DeviceStatusDiscovered

	DeviceTypePhone DeviceType = iota
	DeviceTypeHeadphones
	DeviceTypeLaptop
	DeviceTypeTV
	DeviceTypeController
	DeviceTypeGeneric

	DeviceGlyphConnected  = "󰂱"
	DeviceGlyphLaptop     = ""
	DeviceGlyphPhone      = ""
	DeviceGlyphController = "󰊴"
	DeviceGlyphHeadphones = "󰋋"
	DeviceGlyphTV         = "󰍹"
	DeviceGlyphGeneric    = "󰾰"
)

var deviceGlyphs = map[DeviceType]string{
	DeviceTypeLaptop:     DeviceGlyphLaptop,
	DeviceTypePhone:      DeviceGlyphPhone,
	DeviceTypeController: DeviceGlyphController,
	DeviceTypeHeadphones: DeviceGlyphHeadphones,
	DeviceTypeTV:         DeviceGlyphTV,
	DeviceTypeGeneric:    DeviceGlyphGeneric,
}

type (
	// Device represents a Bluetooth device
	Device struct {
		Name   string
		MAC    string
		Line   string // Original line from bluetoothctl
		Status DeviceStatus
		Type   DeviceType
	}

	DeviceStatus int
	DeviceType   int
)

func NewDevice(name string, mac string, rawLine string, status DeviceStatus, deviceType DeviceType) Device {
	return Device{
		createDeviceName(name, status, deviceType),
		mac,
		rawLine,
		status,
		DeviceTypePhone,
	}
}

func (d Device) Equals(other Device) bool {
	return d.MAC == other.MAC
}

func createDeviceName(name string, status DeviceStatus, deviceType DeviceType) string {
	glyph := getDeviceGlyph(deviceType)
	if status == DeviceStatusConnected {
		glyph = DeviceGlyphConnected
	}

	return fmt.Sprintf("%s  %s", glyph, name)
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

func getDeviceGlyph(deviceType DeviceType) string {
	glyph, ok := deviceGlyphs[deviceType]
	if !ok {
		glyph = DeviceGlyphGeneric
	}
	return glyph
}

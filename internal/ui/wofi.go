// Package ui provides user interface functionality using wofi
package ui

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/birgerrydback/wofi-bluetooth/internal/bluetooth"
)

const (
	// WofiCommand is the base command for launching wofi
	WofiCommand = "wofi -d -i --no-sort -p"

	ActionDisableBluetooth    = "  Disable Bluetooth"
	ActionEnableBluetooth     = "󰂲  Enable Bluetooth"
	ActionEnableDiscoverable  = "  Enable discoverable"
	ActionDisableDiscoverable = "  Disable discoverable"
	ActionEnablePairable      = "󰌺  Enable pairable"
	ActionDisablePairable     = "  Disable pairable"
	ActionScan                = "󱉶  Scan"
	ActionGoBack              = "Back"

	DeviceActionPair       = "󰌺  Pair"
	DeviceActionUnpair     = "  Unpair"
	DeviceActionTrust      = "󱚩  Trust"
	DeviceActionUntrust    = "󱎚  Untrust"
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

// promptMenuOptions displays a wofi menu with the given options and returns the chosen option
func promptMenuOptions(options []string, prompt string) string {
	optionsStr := strings.Join(options, "\n")

	lines := len(options)

	cmdStr := fmt.Sprintf("echo -e '%s' | %s \"%s\" -L %d",
		strings.ReplaceAll(optionsStr, "'", "'\\''"), // Escape single quotes for shell
		WofiCommand,
		prompt,
		(lines + 1))

	cmd := exec.Command("bash", "-c", cmdStr)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}

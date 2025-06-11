// Package ui provides user interface functionality using wofi
package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"bytes"
)

const (
	// WofiCommand is the base command for launching wofi
	WofiCommand = "wofi -d -i -p"
	
	// ConnectedIcon is the icon for connected devices
	ConnectedIcon = "󰂱"
	
	// DisconnectedIcon is the icon for disconnected devices
	DisconnectedIcon = "󰾰"
	
	// Divider is used to separate sections in the menu
	Divider = "---------"
	
	// GoBack is the text for the back option
	GoBack = "Back"
)

// ShowMenu displays a wofi menu with the given options and returns the chosen option
func ShowMenu(options []string, prompt string) string {
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

// Package main provides the entry point for the wofi-bluetooth application
package main

import (
	"github.com/birgerrydback/wofi-bluetooth/internal/bluetooth"
	"github.com/birgerrydback/wofi-bluetooth/internal/ui"
)

func main() {
	bt := bluetooth.NewController()
	ui := ui.NewUI(bt)
	ui.ShowMainMenu()
}

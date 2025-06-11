# wofi-bluetooth

A Bluetooth management tool for Linux using wofi for the user interface.

## Features

- Enable/disable Bluetooth
- Connect/disconnect devices
- Pair/unpair devices
- Trust/untrust devices
- Toggle scanning, pairable, and discoverable states
- Clean and simple wofi-based interface

## Requirements

- Go 1.16 or later
- bluetoothctl
- wofi
- rfkill

## Installation

### From Source

```bash
git clone https://github.com/birgerrydback/wofi-bluetooth.git
cd wofi-bluetooth
go build -o wofi-bluetooth ./cmd/wofi-bluetooth
sudo install -m 755 wofi-bluetooth /usr/local/bin/
```

### Using Nix Flake

Add this to your flake inputs:

```nix
inputs = {
  # ...
  wofi-bluetooth.url = "github:birgerrydback/wofi-bluetooth";
};
```

Then add it to your packages:

```nix
environment.systemPackages = with pkgs; [
  # ...
  inputs.wofi-bluetooth.packages.${system}.default
];
```

## Usage

Simply run:

```bash
wofi-bluetooth
```

This will open a wofi menu with Bluetooth options.

## License

MIT


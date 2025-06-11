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
git clone https://github.com/Betongsuggan/wofi-bluetooth.git
cd wofi-bluetooth
go build -o wofi-bluetooth ./cmd/wofi-bluetooth
sudo install -m 755 wofi-bluetooth /usr/local/bin/
```

### Using Nix Flake

Add this to your flake inputs:

```nix
inputs = {
  # ...
  wofi-bluetooth.url = "github:Betongsuggan/wofi-bluetooth";
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

## TODO List

### Planned Features

- [x] Enable/disable Bluetooth
- [x] Connect/disconnect devices
- [x] Pair/unpair devices
- [x] Trust/untrust devices
- [ ] Toggle scanning, pairable, and discoverable states
- [ ] Improve scanning functionality with automatic device discovery
- [ ] Device type identification (headphones, speakers, keyboards, etc.)
- [ ] Display battery information for connected devices
- [ ] Show signal strength for discovered and connected devices
- [ ] Implement audio profile switching for audio devices
- [ ] Add support for Bluetooth LE devices
- [ ] Implement device filtering options (show only audio devices, etc.)
- [ ] Create configuration file for customizing appearance and behavior
- [ ] Add keyboard shortcuts for common actions
- [ ] Add support for device aliases/nicknames

### Technical Improvements

- [ ] Refactor code for better modularity
- [ ] Add comprehensive error handling
- [ ] Implement logging system
- [ ] Add unit and integration tests
- [ ] Create man page documentation
- [ ] Add localization support
- [ ] Optimize performance for devices with many Bluetooth connections
- [ ] Add support for multiple Bluetooth adapters
- [ ] Use correct binary dependencies from the flake e.g. `bluetoothctl`

Contributions to any of these features are welcome! Feel free to open an issue or submit a pull request.

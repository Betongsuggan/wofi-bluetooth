{
  description = "A Bluetooth management tool for Linux using wofi for the user interface";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "wofi-bluetooth";
          version = "0.1.0";
          src = ./.;

          vendorHash = null; # Will be set to the correct hash on first build

          nativeBuildInputs = [ pkgs.makeWrapper ];

          buildInputs = [
            pkgs.bluez
            pkgs.wofi
            pkgs.util-linux # for rfkill
          ];

          postInstall = ''
            wrapProgram $out/bin/wofi-bluetooth \
              --prefix PATH : ${pkgs.lib.makeBinPath [
                pkgs.bluez
                pkgs.wofi
                pkgs.util-linux
              ]}
          '';

          meta = with pkgs.lib; {
            description = "A Bluetooth management tool for Linux using wofi for the user interface";
            homepage = "https://github.com/birgerrydback/wofi-bluetooth";
            license = licenses.mit;
            #maintainers = with maintainers; [ ];
            platforms = platforms.linux;
          };
        };

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            bluez
            wofi
            util-linux
          ];
        };
      }
    );
}

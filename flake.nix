{
  description = "Kaktovik time TUI — display, convert, timer, stopwatch, and alarm";

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
          pname = "kaktovik";
          version = "0.1.0";

          src = ./go-tui;

          # vendor/ directory is checked in; no network fetch needed.
          vendorHash = null;

          # The Go module lives in go-tui/ so the binary defaults to "go-tui";
          # rename it to the canonical command name.
          postInstall = ''
            mv $out/bin/go-tui $out/bin/kaktovik
          '';

          meta = with pkgs.lib; {
            description = "Kaktovik (Inupiaq base-20) time TUI: clock, converter, timer, stopwatch, alarm";
            homepage = "https://github.com/rfuller25/kaktovik";
            license = licenses.mit;
            maintainers = [ ];
            mainProgram = "kaktovik";
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
            libnotify    # provides notify-send
            pulseaudio   # provides paplay (optional, for alarm sound)
          ];
        };
      }
    )
    //
    {
      # NixOS module — add to your configuration.nix imports
      nixosModules.default = { config, lib, pkgs, ... }:
        let
          cfg = config.programs.kaktovik;
        in
        {
          options.programs.kaktovik = {
            enable = lib.mkEnableOption "Kaktovik time TUI";

            package = lib.mkOption {
              type = lib.types.package;
              default = self.packages.${pkgs.system}.default;
              description = "The kaktovik package to install.";
            };

            enableNotifications = lib.mkOption {
              type = lib.types.bool;
              default = true;
              description = "Install libnotify so kaktovik can send desktop notifications.";
            };
          };

          config = lib.mkIf cfg.enable {
            environment.systemPackages =
              [ cfg.package ]
              ++ lib.optional cfg.enableNotifications pkgs.libnotify;
          };
        };

      # Home Manager module
      homeManagerModules.default = { config, lib, pkgs, ... }:
        let
          cfg = config.programs.kaktovik;
        in
        {
          options.programs.kaktovik = {
            enable = lib.mkEnableOption "Kaktovik time TUI";

            package = lib.mkOption {
              type = lib.types.package;
              default = self.packages.${pkgs.system}.default;
              description = "The kaktovik package to install.";
            };
          };

          config = lib.mkIf cfg.enable {
            home.packages = [ cfg.package pkgs.libnotify ];
          };
        };
    };
}

{
  description = "A homelab homepage written in go";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };
  outputs = inputs @ {flake-utils, ...}:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        pkgs = import inputs.nixpkgs {
          inherit system;
        };
      in rec {
        packages = rec {
          nixhome = pkgs.buildGoModule {
            pname = "nixhome";
            version = "0.0.1";
            src = ./.;
            vendorHash = null;
          };
          default = nixhome;
        };

        apps = rec {
          nixhome = {
            type = "app";
            program = "${packages.nixhome}/bin/nixhome";
          };
          default = nixhome;
        };

        devShells = {
          default = pkgs.mkShell {
            packages = with pkgs; [
              ## golang
              delve
              go-outline
              go
              golangci-lint
              golangci-lint-langserver
              gopkgs
              gopls
              gotools
            ];

            shellHook = ''
              zsh
            '';

            # Need to disable fortify hardening because GCC is not built with -oO,
            # which means that if CGO_ENABLED=1 (which it is by default) then the golang
            # debugger fails.
            # see https://github.com/NixOS/nixpkgs/pull/12895/files
            hardeningDisable = ["fortify"];
          };
        };
      }
    );
}

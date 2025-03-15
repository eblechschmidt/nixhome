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
        lib = pkgs.lib;
        nixhome = pkgs.callPackage ./nix/package.nix {};
        app = {
          type = "app";
          program = "${nixhome}/bin/nixhome";
        };
        module = ./nix/module.nix;
      in {
        packages = {
          nixhome = nixhome;
          default = nixhome;
        };

        nixosModules = {
          nixhome = module;
          default = module;
        };

        apps = {
          nixhome = app;
          default = app;
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
              ${lib.getExe pkgs.zsh}
              exit
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

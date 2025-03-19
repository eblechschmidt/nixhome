{
  description = "A homelab homepage written in go";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };
  outputs = inputs @ {flake-utils, ...}: let
    systems = [
      "x86_64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
      "aarch64-linux"
    ];
    eachSystem = f:
      builtins.listToAttrs (
        builtins.map (system: {
          name = system;
          value = let
            pkgs = inputs.nixpkgs.legacyPackages.${system};
          in
            f {
              pkgs = pkgs;
              lib = pkgs.lib;
              inherit system;
            };
        })
        systems
      );
  in {
    nixosModules = {
      nixhome = ./nix/module.nix;
      default = ./nix/module.nix;
    };
    legacyPackages = eachSystem (
      {
        pkgs,
        lib,
        ...
      }: let
        nixhome = pkgs.callPackage ./nix/package.nix {};
      in {
        nixhome = nixhome;
        default = nixhome;
      }
    );

    apps = eachSystem (
      {
        pkgs,
        lib,
        ...
      }: let
        nixhome = pkgs.callPackage ./nix/package.nix {};
        app = {
          type = "app";
          program = "${nixhome}/bin/nixhome";
        };
      in {
        nixhome = app;
        default = app;
      }
    );

    devShells = eachSystem (
      {
        pkgs,
        lib,
        ...
      }: {
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
      }
    );
  };
}

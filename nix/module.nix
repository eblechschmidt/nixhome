{
  config,
  lib,
  pkgs,
  ...
}: let
  cfg = config.services.nixhome;
  nhconfig = pkgs.writeTextFile (lib.generators.toYAML cfg.settings);
in {
  options = {
    services.nixhome = {
      enable = lib.mkEnableOption "Enable nix home service";
      address = lib.mkOption {
        description = "Address the web server should bind to.";
        type = lib.types.str;
        default = ":8080";
        example = ":8080";
      };
      dataDir = lib.mkOption {
        description = "Location for icon cache and template storage.";
        type = lib.types.str;
        default = "/var/lib/nixhome";
      };
      package = lib.mkOption {
        type = lib.types.package;
        default = pkgs.callPackage ./package.nix {};
        defaultText = lib.literalExpression "pkgs.callPackage ./package.nix {};";
        description = "nixhome package to use.";
      };
      settings = {
        colors = {
          icon = lib.mkOption {
            description = "Background color for dark theme.";
            type = lib.types.str;
            default = "#2E3440";
            example = "#2E3440";
          };
          dark = {
            background = lib.mkOption {
              description = "Background color for dark theme.";
              type = lib.types.str;
              default = "#2E3440";
              example = "#2E3440";
            };
            text = lib.mkOption {
              description = "Normal text color for dark theme.";
              type = lib.types.str;
              default = "#5E81AC";
              example = "#5E81AC";
            };
            accent = lib.mkOption {
              description = "Accent text color for dark theme.";
              type = lib.types.str;
              default = "#8FBCBB";
              example = "#8FBCBB";
            };
          };
          light = {
            background = lib.mkOption {
              description = "Background color for light theme.";
              type = lib.types.str;
              default = "#2E3440";
              example = "#2E3440";
            };
            text = lib.mkOption {
              description = "Normal text color for light theme.";
              type = lib.types.str;
              default = "#5E81AC";
              example = "#5E81AC";
            };
            accent = lib.mkOption {
              description = "Accent text color for light theme.";
              type = lib.types.str;
              default = "#8FBCBB";
              example = "#8FBCBB";
            };
          };
        };
        apps = lib.types.attrsOf (lib.types.submodule ({name, ...}: {
          options = {
            icon = lib.mkOption {
              description = ''
                URL to an icon or id/name for simple Simple Icons or https://www.svgrepo.com.
              '';
              type = lib.types.str;
              default = "";
              example = "353829/grafana";
            };
            name = lib.mkOption {
              description = "Name of the app.";
              type = lib.types.str;
              default = "";
              example = "grafana";
            };
            url = lib.mkOption {
              description = "URL of the app.";
              type = lib.types.str;
              default = "";
              example = "https://grafana.com/";
            };
          };
        }));
        bookmarks = lib.types.attrsOf (lib.types.submodule ({name, ...}: {
          options = {
            name = lib.mkOption {
              description = "Name of the bookmark.";
              type = lib.types.str;
              default = "";
              example = "grafana";
            };
            url = lib.mkOption {
              description = "URL of the bookmark.";
              type = lib.types.str;
              default = "";
              example = "https://grafana.com/";
            };
          };
        }));
      };
    };
  };
  config = lib.mkIf cfg.enable {
    users.groups."nixhome" = {};
    users.users."nixhome" = {
      group = "nixhome";
      isSystemUser = true;
    };
    systemd.tmpfiles.rules = [
      "d ${cfg.dataDir} 750 nixhome nixhome"
    ];
    systemd.services.nixhome = {
      wantedBy = ["multi-user.target"];
      serviceConfig = {
        Type = "Exec";
        ExecStart = ''
          ${lib.getExe cfg.package} \
            --config ${nhconfig} \
            --dataDir ${cfg.dataDir}
        '';
        Restart = "on-failure";
        User = "nixhome";
        Group = "nixhome";
        DynamicUser = true;
        StateDirectory = "photoprism";
        WorkingDirectory = "${cfg.dataDir}";
        RuntimeDirectory = "${cfg.dataDir}";
        ReadWritePaths = [
          cfg.dataDir
        ];

        LockPersonality = true;
        PrivateDevices = true;
        PrivateUsers = true;
        ProtectClock = true;
        ProtectControlGroups = true;
        ProtectHome = true;
        ProtectHostname = true;
        ProtectKernelLogs = true;
        ProtectKernelModules = true;
        ProtectKernelTunables = true;
        RestrictAddressFamilies = [
          "AF_UNIX"
          "AF_INET"
          "AF_INET6"
        ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        SystemCallArchitectures = "native";
        SystemCallFilter = [
          "@system-service"
          "~@setuid @keyring"
        ];
        UMask = "0066";
      };
    };
  };
}

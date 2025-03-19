# nixhome

`nixhome` is a homelab homepage written in go. The design is heavily inspired by
[sui](https://github.com/jeroenpardon/sui) (I would go as far as calling it
shamelessly copied). A unique feature of `nixhome` is that it retrieves svg icons
and applys a color to it so they nicely fit the design.

# NixOS module

## Installation

If you want to run `nixhome` thrugh a service defined in a NixOS module you just
have to add `nixhome` to you inputs.

```
{
  inputs.nixhome.url = "github:eblechschmidt/nixhome/";
  # optional, not necessary for the module
  #inputs.nixhome.inputs.nixpkgs.follows = "nixpkgs";

  outputs = {
    self,
    nixpkgs,
    nixhome,
  }: {
    # change `yourhostname` to your actual hostname
    nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
      # customize to your system
      system = "x86_64-linux";
      modules = [
        ./configuration.nix
        nixhome.nixosModules.nixhome
      ];
    };
  };
}  
```

## Setup

Configuration is done by adding the following code to your NixOS configuration.

```nix
services.nixhome = {
  enable = true;
  # is an attirbute set modeled after the example config. Please refer to the
  # [Configuration](#configuration) for further information
  # settings = {}...
};
```

# Standalone

## Installation

Install `nixhome` using go:

```
  go install github.com/eblechschmidt/nixhome/cmd/nixhome@latest
```

## Usage

All flags have to be set in order for `nixhome` to start. If the configuration
file is not found `nixhome` will try to create it using the example configuration
in the repository.

If run, `nixhome` also creates a template folder in `dataDir` saving the example
`index.tmpl` and `style.css` to it (only if they do not exist).

```
Usage:
  nixhome [flags]

Flags:
      --addr string      address the web server should bind to (e.g. :8080)
      --config string    config file
      --dataDir string   directory where the icon chache is stored
  -h, --help             help for nixhome
```


# Docker

## Installation

Clone the repo locally.

```
  git clone https://github.com/eblechschmidt/nixhome.git
```

Build the image.

```
  docker-compose build
```

Run the image.


```
  docker-compose up -d
```

## Setup

Both the templates and the example config are written to the `./nixhome/config`
folder.

# Configuration

The configuration is read from a `yaml`-file. An example configuration can be
found in [config/example.yaml](./config/example.yaml).

Colors are specified in hex code starting with a #-sign (e.g. `#121212`). Different
colors for light and dark mode can be specified as well as an icon color. The
icon color is used to generate monochrome icons that match the theme.

`nixhome` has two different kinds of structures, "apps" and "bookmarks". Both are
grouped list of links where apps can also contain icons and bookmarks cannot.

When using the dafault template the icons a re colored with the icon color. If
this is not wanted the following line in the `index.tmpl` has to be hanged from

```
{{.ColorizedIcon}}
```

to

```
{{.Icon}}
```

## Icons

For icons only SVG images are allowed. Icons can be specified in three differnt
ways:

1. [simple-icons](https://github.com/simple-icons/simple-icons)
For simple-icons just specify the icon name (without the extension) of the icon
found in the icons folder in the simple-icons repository.

2. [SVGREPO](https://www.svgrepo.com)
For icons available in the svgrepo specify the the number and the name as seen in
the URL (e.g. "353829/grafana" for https://www.svgrepo.com/svg/353829/grafana).

3. Any arbitrary SVG image can be referenced by specifying the whole URL to that
icon (e.g. https://upload.wikimedia.org/wikipedia/commons/b/bb/Gitea_Logo.svg)

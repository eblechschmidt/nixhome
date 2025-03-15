{pkgs, ...}:
pkgs.buildGoModule {
  pname = "nixhome";
  version = "0.1.0";
  src = ../.;
  vendorHash = "sha256-dkVd8xP16abFbmlWbqtkOAJKtH3mpdZXtpIzfq8Lo8M=";
  meta = {
    mainProgram = "nixhome";
  };
}

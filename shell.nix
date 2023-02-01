{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/0218941ea68b4c625533bead7bbb94ccce52dceb.tar.gz") {}
}:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.golangci-lint
    pkgs.nodejs
    pkgs.nodePackages_latest.pnpm
    pkgs.protobuf
    pkgs.protoc-gen-go
  ];
}

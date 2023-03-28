{ pkgs ? import (fetchTarball
  "https://github.com/NixOS/nixpkgs/archive/cfe5833be50b3f0a3c77cb03b43483f139c9ec04.tar.gz")
  { } }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go_1_20
    pkgs.golangci-lint
    pkgs.nodejs-18_x
    pkgs.nodePackages_latest.pnpm
    pkgs.protobuf
    pkgs.protoc-gen-go
    pkgs.nixfmt
    pkgs.goreleaser
  ];
}

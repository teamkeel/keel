{ pkgs ? import (fetchTarball
  "https://github.com/NixOS/nixpkgs/archive/23.05.tar.gz")
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

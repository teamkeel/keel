{ pkgs ? import (fetchTarball
  "https://github.com/NixOS/nixpkgs/archive/4ae2e647537bcdbb82265469442713d066675275.tar.gz")
  { } }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go_1_23
    pkgs.golangci-lint
    pkgs.nodejs-18_x
    pkgs.nodePackages_latest.pnpm
    pkgs.protobuf
    pkgs.protoc-gen-go
    pkgs.nixfmt
    pkgs.goreleaser
  ];
}

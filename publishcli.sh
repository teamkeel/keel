#!/bin/bash

# Compiles the keel cli for OS X + Apple silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/teamkeel/keel/cmd.Version=$1'" -o ./dist/keel cmd/keel/main.go

# compress the binary using upx
sudo apt-get install -y upx
upx --brute dist/keel

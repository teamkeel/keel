#!/bin/bash

# Compiles the keel cli for OS X + Apple silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'github.com/teamkeel/keel/cmd.Version=$1'" -o ./dist/keel cmd/keel/main.go

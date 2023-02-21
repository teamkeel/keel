#!/bin/bash

# Compiles the keel cli for OS X + Apple silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'github.com/teamkeel/keel/cmd.Version=$1'" -o ./dist/keel cmd/keel/main.go

# Set the keel binary as an executable in the git index
# https://git-scm.com/docs/git-update-index#Documentation/git-update-index.txt---chmod-x
git add --chmod=+x dist/keel

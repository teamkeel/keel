#!/bin/bash

go build -ldflags="-X 'github.com/teamkeel/keel/cmd.Version=$1'" -o ./dist/keel cmd/keel/main.go

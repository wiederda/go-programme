#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/MediaFolderRenamer main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/MediaFolderRenamer.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/MediaFolderRenamer main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/MediaFolderRenamer main.go
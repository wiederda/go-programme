#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/updater main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/updater.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/updater main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/updater main.go
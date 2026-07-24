#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/CopyMaster main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/CopyMaster.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/CopyMaster main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/CopyMaster main.go
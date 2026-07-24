#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/archiver main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/archiver.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/archiver main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/archiver main.go
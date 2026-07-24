#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/FindReplace main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/FindReplace.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/FindReplace main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/FindReplace main.go
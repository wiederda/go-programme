#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/get-git main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/get-git.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/get-git main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/get-git main.go
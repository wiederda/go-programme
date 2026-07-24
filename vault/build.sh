#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/vault main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/vault.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/vault main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/vault main.go
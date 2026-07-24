#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/cryptdecrypt main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/cryptdecrypt.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/cryptdecrypt main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/cryptdecrypt main.go
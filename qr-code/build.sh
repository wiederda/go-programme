#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/qr-code main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/qr-code.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/qr-code main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/qr-code main.go


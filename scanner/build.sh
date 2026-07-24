#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/scanner main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/scanner.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/scanner main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/scanner main.go


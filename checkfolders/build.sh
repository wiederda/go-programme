#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/checkfolders main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/checkfolders.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/checkfolders main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/checkfolders main.go
#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/FolderFileMatcher main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/FolderFileMatcher.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/FolderFileMatcher main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/FolderFileMatcher main.go
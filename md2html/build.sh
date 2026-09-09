#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/md2html main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/md2html.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/md2html main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/md2html main.go



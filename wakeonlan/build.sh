#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/wakeonlan main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/wakeonlan.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/wakeonlan main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/wakeonlan main.go


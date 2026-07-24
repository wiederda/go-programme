#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/openssl main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/openssl.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/openssl main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/openssl main.go


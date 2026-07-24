#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/docker-compose-converter main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/docker-compose-converter.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/docker-compose-converter main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/docker-compose-converter main.go


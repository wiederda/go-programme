#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/FastCopy main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/FastCopy.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/FastCopy main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/FastCopy main.go


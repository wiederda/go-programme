#!/bin/bash

mkdir -p ./linux ./windows ./macos/amd64 ./macos/arm64
GO=/usr/bin/go

GOOS=linux GOARCH=amd64 $GO build -o ./linux/filecheck main.go
GOOS=windows GOARCH=amd64 $GO build -o ./windows/filecheck.exe main.go
GOOS=darwin GOARCH=amd64 $GO build -o ./macos/amd64/filecheck main.go
GOOS=darwin GOARCH=arm64 $GO build -o ./macos/arm64/filecheck main.go
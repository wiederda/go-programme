#!/bin/bash

GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=clang go build -o ./linux/CalibreUnsortedFixer main.go
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=clang go build -o ./windows/CalibreUnsortedFixer.exe main.go
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CC=clang go build -o ./macos/amd64/CalibreUnsortedFixer main.go
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 CC=clang go build -o ./macos/arm64/CalibreUnsortedFixer main.go

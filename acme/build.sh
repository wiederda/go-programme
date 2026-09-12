#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./linux/acme .
GOOS=windows GOARCH=amd64 go build -o ./windows/acme.exe .
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/acme .
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/acme .


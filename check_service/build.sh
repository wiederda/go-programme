#!/bin/bash

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./windows/check_service.exe main.go

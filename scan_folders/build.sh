GOOS=linux GOARCH=amd64 go build -o ./linux/scan_folders main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/scan_folders.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/scan_folders main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/scan_folders main.go

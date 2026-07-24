GOOS=linux GOARCH=amd64 go build -o ./linux/CopyFiles main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/CopyFiles.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/CopyFiles main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/CopyFiles main.go

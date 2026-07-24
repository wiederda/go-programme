GOOS=linux GOARCH=amd64 go build -o ./linux/bildtotext main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/bildtotext.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/bildtotext main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/bildtotext main.go

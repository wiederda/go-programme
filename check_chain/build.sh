GOOS=linux GOARCH=amd64 go build -o ./linux/check_chain main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/check_chain.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/check_chain main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/check_chain main.go

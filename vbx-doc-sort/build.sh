GOOS=linux GOARCH=amd64 go build -o ./linux/vbx-doc-sort main.go
GOOS=windows GOARCH=amd64 go build -o ./windows/vbx-doc-sort.exe main.go
GOOS=darwin GOARCH=amd64 go build -o ./macos/amd64/vbx-doc-sort main.go
GOOS=darwin GOARCH=arm64 go build -o ./macos/arm64/vbx-doc-sort main.go
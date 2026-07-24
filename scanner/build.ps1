$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./linux/scanner main.go

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o ./windows/scanner.exe main.go

$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o ./macos/amd64/scanner main.go

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o ./macos/arm64/scanner main.go
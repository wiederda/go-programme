$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./linux/openssl main.go

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o ./windows/openssl.exe main.go

$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o ./macos/amd64/openssl main.go

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o ./macos/arm64/openssl main.go
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./linux/cryptdecrypt main.go

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o ./windows/cryptdecrypt.exe main.go

$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o ./macos/amd64/cryptdecrypt main.go

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o ./macos/arm64/cryptdecrypt main.go
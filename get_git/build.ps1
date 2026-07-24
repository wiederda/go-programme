$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./linux/get-git main.go

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o ./windows/get-git.exe main.go

$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o ./macos/amd64/get-git main.go

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o ./macos/arm64/get-git main.go
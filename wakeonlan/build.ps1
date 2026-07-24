$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./linux/wakeonlan main.go

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o ./windows/wakeonlan.exe main.go

$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o ./macos/amd64/wakeonlan main.go

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o ./macos/arm64/wakeonlan main.go
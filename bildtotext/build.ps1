$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./linux/bildtotext main.go

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o ./windows/bildtotext.exe main.go

$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o ./macos/amd64/bildtotext main.go

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o ./macos/arm64/bildtotext main.go
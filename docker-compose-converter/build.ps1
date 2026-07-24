$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./linux/docker-compose-converter main.go

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o ./windows/docker-compose-converter.exe main.go

$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o ./macos/amd64/docker-compose-converter main.go

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o ./macos/arm64/docker-compose-converter main.go
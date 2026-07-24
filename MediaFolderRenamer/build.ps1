$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./linux/MediaFolderRenamer main.go

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o ./windows/MediaFolderRenamer.exe main.go

$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o ./macos/amd64/MediaFolderRenamer main.go

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o ./macos/arm64/MediaFolderRenamer main.go
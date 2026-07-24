$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ./linux/FolderFileMatcher main.go

$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o ./windows/FolderFileMatcher.exe main.go

$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o ./macos/amd64/FolderFileMatcher main.go

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o ./macos/arm64/FolderFileMatcher main.go
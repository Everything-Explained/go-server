
.DEFAULT_GOAL := build

.EXEC:fmt vet tidy build lint
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

tidy: vet
	go mod tidy

build: tidy
	go build -o "bin/server.exe" "cmd/server/main.go"

lint:
	perfsprint --fix ./...
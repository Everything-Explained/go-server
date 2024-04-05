
.DEFAULT_GOAL := build

.EXEC:fmt vet tidy build lint
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

tidy: vet
	go mod tidy

build: tidy
	go build -o server.exe

lint:
	perfsprint --fix ./...
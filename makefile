
.DEFAULT_GOAL := build

.EXEC:fmt vet tidy build lint
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

tidy: fmt
	go mod tidy

build: tidy
	go build -o server.exe

lint:
	perfsprint --fix ./...
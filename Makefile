.PHONY: build run clean lint fmt modernize test fix smoke-docs integration-test wsl

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags "-X main.version=$(VERSION)" .

run: build
	./fz

clean:
	rm -f fz

lint:
	golangci-lint-v2 run ./...

test:
	go test ./...

fix:
	go fix ./...

fmt:
	gofumpt -l -w .

modernize:
	go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest -fix ./...

smoke-docs:
	bash script/docs_smoke_test.sh

integration-test: build
	bash script/integration_test.sh

wsl:
	go install github.com/bombsimon/wsl/v5/cmd/wsl@main
	wsl --fix ./...
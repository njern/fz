.PHONY: build run clean lint fmt fmt-check modernize test fix smoke-docs integration-test \
	install-tools wsl wsl-check ci pre-commit fix-all tools-versions \
	check-golangci-lint check-gofumpt check-modernize check-wsl

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GOLANGCI_LINT_VERSION ?= v2.11.3
GOFUMPT_VERSION ?= v0.9.2
MODERNIZE_VERSION ?= v0.40.0
WSL_VERSION ?= v5.0.0

check-golangci-lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint is required for 'make lint'."; \
		echo "Run 'make install-tools' to install the required developer tools."; \
		echo "Or install just this tool and ensure it is on PATH."; \
		echo "Example: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)"; \
		exit 1; \
	}

check-gofumpt:
	@command -v gofumpt >/dev/null 2>&1 || { \
		echo "gofumpt is required for 'make fmt'."; \
		echo "Run 'make install-tools' to install the required developer tools."; \
		echo "Or install just this tool and ensure it is on PATH."; \
		echo "Example: go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)"; \
		exit 1; \
	}

check-wsl:
	@command -v wsl >/dev/null 2>&1 || { \
		echo "wsl is required for 'make wsl'."; \
		echo "Run 'make install-tools' to install the required developer tools."; \
		echo "Or install just this tool and ensure it is on PATH."; \
		echo "Example: go install github.com/bombsimon/wsl/v5/cmd/wsl@$(WSL_VERSION)"; \
		exit 1; \
	}

check-modernize:
	@command -v modernize >/dev/null 2>&1 || { \
		echo "modernize is required for 'make modernize'."; \
		echo "Run 'make install-tools' to install the required developer tools."; \
		echo "Or install just this tool and ensure it is on PATH."; \
		echo "Example: go install golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@$(MODERNIZE_VERSION)"; \
		exit 1; \
	}

build:
	go build -ldflags "-X main.version=$(VERSION)" .

run: build
	./fz

clean:
	rm -f fz

install-tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	go install golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@$(MODERNIZE_VERSION)
	go install github.com/bombsimon/wsl/v5/cmd/wsl@$(WSL_VERSION)

tools-versions:
	@echo "golangci-lint $(GOLANGCI_LINT_VERSION)"
	@echo "gofumpt $(GOFUMPT_VERSION)"
	@echo "modernize $(MODERNIZE_VERSION)"
	@echo "wsl $(WSL_VERSION)"

lint: check-golangci-lint
	golangci-lint run ./...

test:
	go test ./...

fix:
	go fix ./...

fmt: check-gofumpt
	gofumpt -l -w .

fmt-check: check-gofumpt
	@out="$$(gofumpt -l .)"; \
	if [ -n "$$out" ]; then \
		echo "$$out"; \
		echo "Code is not gofumpt-formatted. Run 'make fmt'."; \
		exit 1; \
	fi

modernize: check-modernize
	modernize -fix ./...

smoke-docs:
	bash script/docs_smoke_test.sh

integration-test: build
	bash script/integration_test.sh

wsl: check-wsl
	wsl --fix ./...

wsl-check: check-wsl
	go list ./... | xargs -n 1 wsl

ci: lint test smoke-docs fmt-check wsl-check

pre-commit: ci

fix-all: fix fmt wsl

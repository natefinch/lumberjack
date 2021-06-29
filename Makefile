BINDIR := $(CURDIR)/bin

HAS_GOLANGCI := $(shell command -v golangci-lint;)
HAS_AIR := $(shell command -v air;)

build:
	@GOBIN=$(BINDIR) go install -race && echo "Build OK"

dependencies-dev:
	@echo "Development dependencies and versions:"
	@echo "- GolangCI Linter  1.41.x"
	@echo "- Air 1.15.x"

lint:
ifndef HAS_GOLANGCI
	$(error You must install github.com/golangci/golangci-lint)
endif
	@golangci-lint run -v -c .golangci.yml && echo "Lint OK"

test:
	@go test -p 1 -cover -coverprofile=coverage.out -run . ./... && echo "Test OK"

coverage: test
	@go tool cover -func=coverage.out && echo "Coverage OK"

clean:
	@go clean ./...
	@rm -rf $(BINDIR)
	@rm -f coverage.*

dev:
ifndef HAS_AIR
	$(error You must install github.com/cosmtrek/air)
endif
	@air -c .air.toml || (make dependencies-dev; exit 1)

ci: coverage lint build
cli: build

.PHONY: build lint test coverage clean dev
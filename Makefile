HAS_GOLANGCI := $(shell command -v golangci-lint;)

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

ci: lint test coverage

.PHONY: lint test coverage clean ci
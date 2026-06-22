PKG := "aipm"
PKG_LIST := $(shell go list ${PKG}/... | grep -v '/api-tests')

.DEFAULT_GOAL := check
check: lint vet race ## Check project

init:
	@go install golang.org/x/lint/golint@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest

lint: ## Lint the files
	@staticcheck ${PKG_LIST}
	@golint -set_exit_status ${PKG_LIST}

vet: ## Vet the files
	@go vet ${PKG_LIST}

test: ## Run tests
	@go test -short ${PKG_LIST}

api-test: ## Run integration tests
	@go build -o aipm-server cmd/server/*.go
	@PORT=18080 ./aipm-server >/tmp/aipm-api-test.log 2>&1 & pid=$$!; \
		trap 'kill $$pid >/dev/null 2>&1 || true; rm -f aipm-server' EXIT; \
		for _ in $$(seq 1 50); do \
			if curl -fsS http://localhost:18080/api/v1/health >/dev/null 2>&1; then \
				AIPM_API_BASE_URL=http://localhost:18080 go test -count=1 ./api-tests/...; \
				exit $$?; \
			fi; \
			sleep 0.1; \
		done; \
		cat /tmp/aipm-api-test.log; \
		exit 1

race: ## Run tests with data race detector
	@go test -race ${PKG_LIST}

benchmark: ## Run benchmarks
	@go test -run="-" -benchmem -bench=".*" ${PKG_LIST}

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

goversion ?= "1.25"
test_version: ## Run tests inside Docker with given version (defaults to 1.25 oldest supported). Example: make test_version goversion=1.25
	@docker run --rm -it -v $(shell pwd):/project golang:$(goversion) /bin/sh -c "cd /project && make init check"

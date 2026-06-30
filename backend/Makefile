PKG := "aipm"
PKG_LIST := $(shell go list ${PKG}/... | grep -v '/api-tests')
API_TEST_DB_URL ?= postgres://postgres:postgres@localhost:5432/changes_test
API_TEST_PORT ?= 19080
API_TEST_BASE_URL ?= http://localhost:$(API_TEST_PORT)

.DEFAULT_GOAL := check
check: lint vet race ## Check project

init:
	@go install golang.org/x/lint/golint@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest

lint: ## Lint the files
	@goimports -w .
	@staticcheck ${PKG_LIST}
	@golint -set_exit_status ${PKG_LIST}

vet: ## Vet the files
	@go vet ${PKG_LIST}

test: ## Run tests
	@go test -short ${PKG_LIST}

api-test: ## Rebuild changes_test and run HTTP integration tests
	@pkill -9 -x aipm-server >/dev/null 2>&1 || true
	@go clean -testcache
	@case "$(API_TEST_DB_URL)" in \
		*/changes_test|*/changes_test\?*) ;; \
		*) echo "API_TEST_DB_URL must target changes_test"; exit 1 ;; \
	esac
	@psql $(API_TEST_DB_URL) -v ON_ERROR_STOP=1 -f ../db/init.sql >/tmp/aipm-api-test-db.log 2>&1
	@psql $(API_TEST_DB_URL) -v ON_ERROR_STOP=1 -f ../db/seed.sql >>/tmp/aipm-api-test-db.log 2>&1
	@go build -o aipm-server cmd/server/*.go
	@./aipm-server -db $(API_TEST_DB_URL) -port $(API_TEST_PORT) >/tmp/aipm-api-test.log 2>&1 & pid=$$!; \
		trap 'kill $$pid >/dev/null 2>&1 || true; rm -f aipm-server' EXIT; \
		for _ in $$(seq 1 50); do \
			if curl -fsS $(API_TEST_BASE_URL)/api/v1/health >/dev/null 2>&1; then \
				API_TEST_BASE_URL=$(API_TEST_BASE_URL) go test -count=1 ./api-tests/...; \
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

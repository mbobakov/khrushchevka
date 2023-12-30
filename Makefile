.DEFAULT_GOAL := help

BIN_DIR = ${PWD}/bin

export GOPRIVATE=github.com/reemote/*
export PATH := ${BIN_DIR}:$(PATH)

BOLD=$(shell tput -T xterm bold)
RED=$(shell tput -T xterm setaf 1)
GREEN=$(shell tput -T xterm setaf 2)
YELLOW=$(shell tput -T xterm setaf 3)
RESET=$(shell tput -T xterm sgr0)


.PHONY: help
help: ## Display help screen
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
	{printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: tools
tools: ## Installing tools from tools.go
	- echo Installing tools from tools.go
	- cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % env GOBIN=${BIN_DIR} go install %

.PHONY: clean
clean: ## run all cleanup tasks
	go clean ./...
	rm -rf $(BIN_DIR)

.PHONY: test
test: generate ## Run unit tests
	go test -v ./... -count 1 -race --failfast
	@echo ""
	@echo "${GREEN} All tests passed ✅"
	@echo "${RESET}"

test-integration:  ## Integration test
test-integration:
	go test -timeout 300s -tags integration -count 1 -v ./...
	@echo ""
	@echo "${GREEN} All tests passed ✅"
	@echo "${RESET}"

.PHONY: lint
lint: generate tools ## Run linter
	${BIN_DIR}/golangci-lint --color=always run ./... -v --timeout 5m

.PHONY: generate
generate: tools ## Go Generate
	go generate ./...


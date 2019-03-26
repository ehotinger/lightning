.PHONY: build
build: ## Build the Go packages
	@echo "+ $@"
	@go build ./...

.PHONY: test
test: ## Runs the Go tests
	@echo "+ $@"
	@go test ./...

.PHONY: help
help: ## Prints this help menu
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort	
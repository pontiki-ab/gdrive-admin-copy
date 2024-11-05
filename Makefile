MAIN_FILE=cmd/nnx/nnx.go


.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: lint
lint: assert_golangci_lint_installed ## Run code linters.
	golangci-lint run ./... --concurrency 2 -v -c .golangci.yaml

.PHONY: lint_docker
lint_docker: assert_docker_installed ## Run code linters in a container (slow but reproducible).
	docker image pull golangci/golangci-lint:latest-alpine
	docker run -v .:/src -w "/src" golangci/golangci-lint:latest-alpine golangci-lint run ./... --concurrency 2 -v -c .golangci.yaml

# there's no official docker image for this, so if you want to use gofumpt, you need to install it locally (for now)
.PHONY: fumpt
fumpt: assert_gofumpt_installed ## Run gofumpt to fix all formatting issues in project.
	gofumpt -w .

.PHONY: build
build: assert_go_installed ## Build binary.
	GOOS=darwin GOARCH=arm64 go build -o nnx.macos.arm64 $(MAIN_FILE)
	chmod +x nnx.macos.arm64
	GOOS=darwin GOARCH=amd64 go build -o nnx.macos.amd64 $(MAIN_FILE)
	chmod +x nnx.macos.amd64

.PHONY: build_docker
build_docker: assert_docker_installed ## Build binary in a container (slow but reproducible).
	docker image pull golang:1-alpine
	docker run -v .:/src -w "/src" -e GOOS=darwin -e GOARCH=arm64 golang:1-alpine go build -o nnx.macos.arm64 $(MAIN_FILE)
	chmod +x nnx.macos.arm64
	docker run -v .:/src -w "/src" -e GOOS=darwin -e GOARCH=amd64 golang:1-alpine go build -o nnx.macos.amd64 $(MAIN_FILE)
	chmod +x nnx.macos.amd64

.PHONY: mod
mod: assert_go_installed ## Update go modules.
	go get -u -t ./...
	go mod tidy
	go mod vendor

.PHONY: mod_docker
mod_docker: assert_docker_installed ## Update go modules (slow but reproducible).
	docker image pull golang:1-alpine
	docker run -v .:/src -w "/src" golang:1-alpine go get -u -t ./...
	docker run -v .:/src -w "/src" golang:1-alpine go mod tidy
	docker run -v .:/src -w "/src" golang:1-alpine go mod vendor

.PHONY: install
install: assert_go_installed ## Install binary locally.
	go install $(MAIN_FILE)


##@ Assertions
.PHONY: assert_go_installed
assert_go_installed: ## Assert go is installed.
	@if ! command -v go &> /dev/null; then \
		echo "go is not installed; you need to install it in order to run this command"; \
		exit 1; \
	fi

.PHONY: assert_docker_installed
assert_docker_installed: ## Assert docker is installed.
	@if ! command -v docker &> /dev/null; then \
		echo "docker is not installed; you need to install it in order to run this command"; \
		exit 1; \
	fi

.PHONY: assert_golangci_lint_installed
assert_golangci_lint_installed: ## Assert golangci-lint is installed.
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint is not installed; you need to install it in order to run this command"; \
		exit 1; \
	fi

.PHONY: assert_gofumpt_installed
assert_gofumpt_installed: ## Assert gofumpt is installed.
	@if ! command -v gofumpt &> /dev/null; then \
		echo "gofumpt is not installed; you need to install it in order to run this command"; \
		exit 1; \
	fi

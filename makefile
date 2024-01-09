# Set ENV to dev, prod, etc. to load .env.$(ENV) file
ENV ?= 
-include .env
export
-include .env.$(ENV)
export

# Internal variables you don't want to change
SHELL := /bin/bash
SRC_DIR := ./cmd
REPO_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
GOLINT_PATH := $(REPO_DIR)/.tools/golangci-lint
AIR_PATH := $(REPO_DIR)/.tools/air

.EXPORT_ALL_VARIABLES:
.PHONY: help image push build run lint lint-fix
.DEFAULT_GOAL := help

# Override these if building your own images
IMAGE_REG ?= ghcr.io
IMAGE_NAME ?= benc-uk/mockery
IMAGE_TAG ?= latest
IMAGE_PREFIX := $(IMAGE_REG)/$(IMAGE_NAME)

help: ## 💬 This help message :)
	@figlet Mockery || true
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tools: ## 🔮 Install dev tools into project .tools directory
	@figlet $@ || true
	@$(GOLINT_PATH) > /dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./.tools/
	@$(AIR_PATH) -v > /dev/null 2>&1 || ( wget https://github.com/cosmtrek/air/releases/download/v1.42.0/air_1.42.0_linux_amd64 -q -O .tools/air && chmod +x .tools/air )
	
lint: ## 🔍 Lint & format check only, sets exit code on error for CI
	@figlet $@ || true
	cd $(SRC_DIR); $(GOLINT_PATH) run --timeout 3m

lint-fix: ## 📝 Lint & format, attempts to fix errors & modify code
	@figlet $@ || true
	cd $(SRC_DIR); $(GOLINT_PATH) run --timeout 3m --fix

image: check-vars ## 📦 Build container image from Dockerfile
	@figlet $@ || true
	docker build --file ./build/Dockerfile \
	--tag $(IMAGE_PREFIX):$(IMAGE_TAG) . 

push: check-vars ## 📤 Push container image to registry
	@figlet $@ || true
	docker push $(IMAGE_PREFIX):$(IMAGE_TAG)

build: ## 🔨 Build binaries for all platforms
	@figlet $@ || true
	GOOS=linux GOARCH=amd64 go build -o bin/mockery-linux $(SRC_DIR)/...
	GOOS=windows GOARCH=amd64 go build -o bin/mockery-windows $(SRC_DIR)/...
	GOOS=darwin GOARCH=arm64 go build -o bin/mockery-mac $(SRC_DIR)/...

run: ## 🏃 Build and test the CLI, and hot reload on changes
	@figlet $@ || true
	$(AIR_PATH) -c .air.toml

clean: ## 🧹 Clean up, remove dev data and files
	@figlet $@ || true
	@rm -rf bin .tools tmp

release: build ## 🚀 Release a new version on GitHub
	@figlet $@ || true
	@if [[ -z "${VERSION}" ]]; then echo "💥 Error! Required variable VERSION is not set!"; exit 1; fi
	@echo "Releasing version $(VERSION) on GitHub"
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	gh release create "$(VERSION)" --title "v$(VERSION)" \
	--latest --notes "Release v$(VERSION)" 
	gh release upload "$(VERSION)" ./bin/mockery-linux ./bin/mockery-windows ./bin/mockery-mac

test: ## 🧪 Run unit tests
	@figlet $@ || true
	go test -v ./...
	
check-vars:
	@if [[ -z "${IMAGE_REG}" ]]; then echo "💥 Error! Required variable IMAGE_REG is not set!"; exit 1; fi
	@if [[ -z "${IMAGE_NAME}" ]]; then echo "💥 Error! Required variable IMAGE_NAME is not set!"; exit 1; fi
	@if [[ -z "${IMAGE_TAG}" ]]; then echo "💥 Error! Required variable IMAGE_TAG is not set!"; exit 1; fi

install: build ## 💽 Build and install the CLI
	@figlet $@ || true
	cp ./bin/mockery-linux ~/.local/bin/mockery

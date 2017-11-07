
SOURCE_VERSION = $(shell git rev-parse --short=6 HEAD)
BUILD_FLAGS = -v -ldflags "-X github.com/mdevilliers/lambda-deployer.SourceVersion=$(SOURCE_VERSION)"
PACKAGES := $(shell go list ./... | grep -v /vendor/ )

GO_TEST = go test -covermode=atomic
GO_INTEGRATION = $(GO_TEST) -bench=. -v --tags=integration
GO_COVER = go tool cover
GO_BENCH = go test -bench=.
ARTEFACT_DIR = coverage

all: linux-arm linux-amd64 darwin-amd64 ## build executables for the various environments

.PHONY: all

get-build-deps: ## install build dependencies
	# for checking licences
	go get github.com/chespinoza/goliscan

.PHONY: get-build-deps

check-vendor-licenses: ## check if licenses of project dependencies meet project requirements
	@goliscan check --direct-only -strict
	@goliscan check --indirect-only -strict

.PHONY: check-vendor-licenses

test: ## run tests
	$(GO_TEST) $(PACKAGES)

.PHONY: test

test_integration: ## run integration tests (SLOW)
	mkdir -p $(ARTEFACT_DIR)
	echo 'mode: atomic' > $(ARTEFACT_DIR)/cover-integration.out
	touch $(ARTEFACT_DIR)/cover.tmp
	$(foreach package, $(PACKAGES), $(GO_INTEGRATION) -coverprofile=$(ARTEFACT_DIR)/cover.tmp $(package) && tail -n +2 $(ARTEFACT_DIR)/cover.tmp >> $(ARTEFACT_DIR)/cover-integration.out || exit;)
.PHONY: test_integration

clean: ## clean up
	rm -rf tmp/
	rm -rf $(ARTEFACT_DIR)

.PHONY: clean

bench: ## run benchmark tests
	$(GO_BENCH) $(PACKAGES)

.PHONY: bench

coverage: test_integration ## generate and display coverage report
	$(GO_COVER) -func=$(ARTEFACT_DIR)/cover-integration.out

.PHONY: test_integration

lambda-build: ## build the lambda zip files
	cd ./cmd/deployer/ && $(MAKE)

.PHONY: lambda-build

lambda-build-ci: ## build the lambda zip files (using the ci build)
	cd ./cmd/deployer/ && $(MAKE) ci

.PHONY: lambda-build

darwin-amd64: tmp/build/darwin-amd64/lambda-uploader ## build for mac amd64

linux-amd64: tmp/build/linux-amd64/lambda-uploader ## build for linux amd64

linux-arm: tmp/build/linux-arm/lambda-uploader ## build for linux arm (raspberry-pi)

.PHONY: darwin-amd64 linux-amd64 linux-arm


## linux-amd64
tmp/build/linux-amd64/lambda-uploader:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(@) ./cmd/uploader/


## linux-arm
tmp/build/linux-arm/lambda-uploader:
	GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o $(@) ./cmd/uploader/


## darwin-amd64
tmp/build/darwin-amd64/lambda-uploader:
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(@) ./cmd/uploader/


# 'help' parses the Makefile and displays the help text
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: help

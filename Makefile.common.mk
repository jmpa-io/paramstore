#
# This Makefile is generic across all our DevEx repositories.
# Any changes made here should also be updated in:
# https://github.com/myob-ops/devex-template/blob/main/Makefile.common.mk
#

# Check project variables are given.
ifndef PROJECT
$(error PROJECT not defined, missing Makefile?)
endif

# Vars.
SHELL			= /bin/sh # The shell used when make runs `shell`.
ENVIRONMENT		?= dev # The environment being deployed; affects what type of resources are used when deploying.
COMMIT			= $(shell git describe --tags --always) # The git commit associated with the repository.
# ---
CMD_SOURCE		= $(shell find cmd/$* -type f) # The source for a command.
BINARIES 		= $(patsubst %,dist/%,$(shell find cmd/* -maxdepth 0 -type d -exec basename {} \;)) # A list of ALL binaries in the repository.
BINARIES_LINUX	= $(patsubst %,%-linux,$(BINARIES)) # A list of ALL linux binaries in the repository.

# Check dependencies (only in the pipeline).
ifndef BUILDKITE
EXECUTABLES ?= awk column grep
$(foreach bin,$(EXECUTABLES),$(if $(shell command -v $(bin) 2> /dev/null),,$(error Please install `$(bin)`)))
endif

# The default command executed when running `make`.
.DEFAULT_GOAL:=	help

# Clear default suffix rules.
.SUFFIXES:

# ┬  ┬ ┌┐┌┌┬┐
# │  │ │││ │
# ┴─┘┴ ┘└┘ ┴

.PHONY: lint
lint: lint-go lint-sh ## ** Lints everything.

.PHONY: lint-go
lint-go: ## Lints Go files.
	@test -z "$(BUILDKITE)" || echo "~~~ :golang: Linting Go."
	golangci-lint run --timeout 5m0s --skip-dirs ./vendor -v --fast -j 15 --allow-parallel-runners

.PHONY: lint-sh
lint-sh: ## Lints shell files.
	@test -z "$(BUILDKITE)" || echo "~~~ :shell: Linting sh."
	find . -name "*.sh" ! -path '*/_mod/*' -exec shellcheck '{}' \;

# ┌┬┐┌─┐┌─┐┌┬┐
#  │ ├┤ └─┐ │
#  ┴ └─┘└─┘ ┴

.PHONY: test
test: test-go ## ** Tests everything.

.PHONY: test-go
test-go: export AWS_XRAY_SDK_DISABLED=TRUE
test-go: dist ## Tests Go + generates coverage report.
	@test -z "$(BUILDKITE)" || echo "~~~ :golang: Unit tests."
	@go version
	go test -short -coverprofile=dist/coverage.txt -covermode=atomic -race -vet=off ./... \
		&& go tool cover -func=dist/coverage.txt


# ┌┐ ┬┌┐┌┌─┐┬─┐┬┌─┐┌─┐
# ├┴┐││││├─┤├┬┘│├┤ └─┐
# └─┘┴┘└┘┴ ┴┴└─┴└─┘└─┘

.PHONY: binaries list-binaries
binaries: $(BINARIES) ## ** Builds ALL binaries.
list-binaries: ## Lists ALL binaries.
	@echo $(BINARIES)
list-binaries-linux: ## Lists ALL Linux binaries.
	@echo $(BINARIES_LINUX)

dist: ; mkdir dist # Creates the dist directory, if missing.

## Builds the binary for the given service (for the local machine platform - GOOS).
binary-%: dist/%
dist/%: $(CMD_SOURCE)
	@test -z "$(BUILDKITE)" || echo "~~~ :golang: Building $@"
	@go version
	go build --trimpath -ldflags "-w -s \
		    -X version.Version=$(COMMIT)" \
		    -o $@ ./cmd/$*

## Builds the binary for the given service (for Linux).
binary-%-linux: dist/%-linux
dist/%-linux: $(CMD_SOURCE)
	@test -z "$(BUILDKITE)" || echo "~~~ :golang: Building $@"
	@go version
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --trimpath -ldflags "-w -s \
		    -X version.Version=$(COMMIT)" \
		    -o $@ ./cmd/$*

# ┌┬┐┬┌─┐┌─┐
# ││││└─┐│
# ┴ ┴┴└─┘└─┘

.PHONY: update-template
update-template: ## Pulls changes from obs-template into this repository.
	git fetch template
	git merge template/master --allow-unrelated-histories

.PHONY: clean
clean: ## Removes generated files and folders, resetting this repository back to its initial clone state.
	@test -z "$(BUILDKITE)" || echo "~~~ :adhesive_bandage: Cleaning up."
	rm -f coverage.* traces.*
	rm -rf dist

.PHONY: help
help: ## Prints this help page.
	$(info Available targets:)
	@awk '/^[a-zA-Z\-\_0-9%\/]+:/ { \
		nb = sub( /^## /, "", helpMsg ); \
		if(nb == 0) { helpMsg = $$0; nb = sub( /^[^:]*:.* ## /, "", helpMsg ); } \
		if (nb) print $$1 helpMsg; \
	} { helpMsg = $$0 }' $(MAKEFILE_LIST) \
	| column -ts: \
	| grep --color '^[^ ]*'

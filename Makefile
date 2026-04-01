LDFLAGS += -X main.version=$$(git describe --always --abbrev=40 --dirty)
TEST?=$$(go list ./... |grep -v 'vendor')
PKG_NAME=ironic
TERRAFORM_PLUGINS=$(HOME)/.terraform.d/plugins
BIN_DIR := $(shell pwd)/bin
GOLANGCI_LINT_BIN := $(BIN_DIR)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.64.8

ifeq ("$(IRONIC_ENDPOINT)", "")
	IRONIC_ENDPOINT := http://127.0.0.1:6385/
	export IRONIC_ENDPOINT
endif

default: build

build:
	go build -ldflags "${LDFLAGS}" -tags "${TAGS}"

install: default
	mkdir -p ${TERRAFORM_PLUGINS}
	mv terraform-provider-ironic ${TERRAFORM_PLUGINS}

fmt:
	gofmt -s -d -e ./ironic

lint: $(GOLANGCI_LINT_BIN)
	GOLANGCI_LINT_CACHE=/tmp/terraform-provider-ironic/golangci-lint-cache/ $(GOLANGCI_LINT_BIN) run ironic

$(GOLANGCI_LINT_BIN):
	mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

test:
	go test -tags "${TAGS}" -v ./ironic

acceptance:
	TF_ACC=true go test -tags "acceptance" -v ./ironic/...

clean:
	rm -f terraform-provider-ironic

.PHONY: build install test fmt lint

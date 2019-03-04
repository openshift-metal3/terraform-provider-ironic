LDFLAGS += -X main.version=$$(git describe --always --abbrev=40 --dirty)
TEST?=$$(go list ./... |grep -v 'vendor')
PKG_NAME=ironic

default: fmt lint build

build:
	go build -ldflags "${LDFLAGS}"

install:
	go install -ldflags "${LDFLAGS}"

fmt:
	go fmt ./ironic .

lint:
	go run golang.org/x/lint/golint -set_exit_status ./ironic .

test:
	go test -v ./ironic

clean:
	rm -f terraform-provider-ironic

.PHONY: build install test fmt lint

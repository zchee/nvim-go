PKG := github.com/zchee/go-xdgbasedir
VERSION := $(shell cat VERSION.txt)

GO_TEST ?= go test
GO_TEST_TARGET ?= .
GO_TEST_PACKAGE ?= ./...
BUILD_IMAGE ?= golang:1.10.2-stretch

.PHONY: test
test:  ## Run the go test
	${GO_TEST} -v -race -run=${GO_TEST_TARGET} ${GO_TEST_PACKAGE}

.PHONY: test.docker
test.docker:  ## Run the go test in the container
	docker run --rm -it -v ${CURDIR}:/go/src/${PKG} ${BUILD_IMAGE} go test -v -race -run=${GO_TEST_TARGET} ${GO_TEST_PACKAGE}


.PHONY: lint
lint: lint.fmt lint.golint lint.vet  ## Run gofmt, go tool vet and golint lint tools

.PHONY: lint.fmt
lint.fmt:
	gofmt -s -l -w .

.PHONY: lint.fmt
lint.vet:
	go vet -v -all -shadow .

.PHONY: lint.golint
lint.golint: $(shell command -v golint)
	golint -min_confidence=0.8 -set_exit_status ./...


.PHONY: help
help:  ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[33m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z.\/_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := test

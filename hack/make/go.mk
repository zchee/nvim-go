# ----------------------------------------------------------------------------
# global

# base
SHELL := /usr/bin/env bash

GO_PATH ?= $(shell go env GOPATH)
GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)

# pkg
PKG = $(subst $(GO_PATH)/src/,,$(CURDIR))
GO_PKGS := $(shell go list ./... | grep -v -e '.pb.go' -e 'api/gate')
GO_APP_PKGS := $(shell go list -f '{{if and (or .GoFiles .CgoFiles) (ne .Name "main")}}{{.ImportPath}}{{end}}' ${PKG}/...)
GO_TEST_PKGS := $(shell go list -f='{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)
GO_VENDOR_PKGS := $(shell go list -f '{{if and (or .GoFiles .CgoFiles) (ne .Name "main")}}./vendor/{{.ImportPath}}{{end}}' ./vendor/...)

GO_TEST ?= go test
GO_TEST_FUNC ?= .
GO_TEST_FLAGS ?=
GO_BENCH_FUNC ?= .
GO_BENCH_FLAGS ?= -benchmem

VERSION=$(shell cat VERSION.txt)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_UNTRACKED_CHANGES=$(shell git status --porcelain --untracked-files=no)
ifneq ($(GIT_UNTRACKED_CHANGES),)
	GIT_COMMIT := $(GIT_COMMIT)-dirty
endif
CTIMEVAR=-X $(PKG)/pkg/version.Tag=$(VERSION) -X $(PKG)/pkg/version.GitCommit=$(GIT_COMMIT)

CGO_ENABLED ?= 0
GO_LDFLAGS=-s -w $(CTIMEVAR)
GO_LDFLAGS_STATIC=-s -w $(CTIMEVAR)
ifneq (${GO_OS},darwin)
	GO_LDFLAGS_STATIC+='-extldflags=-static'
endif

GO_BUILDTAGS=osusergo netgo
GO_BUILDTAGS_STATIC=static static_build
GO_FLAGS ?= -tags='$(GO_BUILDTAGS)' -ldflags="${GO_LDFLAGS}"
GO_INSTALLSUFFIX_STATIC=netgo

ifneq ($(wildcard go.mod),)  # exist go.mod
ifneq ($(GO111MODULE),auto)
	GO_FLAGS+=-mod=vendor
endif
endif

IMAGE_REGISTRY := gcr.io/container-image

# ----------------------------------------------------------------------------
# defines

GOPHER = ""
define target
@printf "$(GOPHER)  \\033[32m$(patsubst ,$@,$(1))\\033[0m\\n"
endef

# ----------------------------------------------------------------------------
# targets

.PHONY: bin/$(APP)
bin/$(APP):
	$(call target,$@)
	CGO_ENABLED=$(CGO_ENABLED) go build -v $(strip $(GO_FLAGS)) -o $@ $(CMD)

bin/$(APP)-race:
	$(call target,$@)
	CGO_ENABLED=$(CGO_ENABLED) go build -v $(strip $(GO_FLAGS)) -o $@ $(CMD)

.PHONY: $(APP)
$(APP): bin/$(APP)

.PHONY: build
build: $(APP)  ## Builds a dynamic executable or package.

.PHONY: build/race
build/race: GO_FLAGS+=-race
build/race: GO_FLAGS+=-ldflags="${GO_LDFLAGS}"
build/race: clean bin/$(APP)-race  ## Build the nvim-go binary with race

.PHONY: static
static: GO_LDFLAGS=${GO_LDFLAGS_STATIC}
static: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
static: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
static: $(APP)  ## Builds a static executable or package.

.PHONY: install
install: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
install: GO_LDFLAGS=${GO_LDFLAGS_STATIC}
install: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
install:  ## Installs the executable or package.
	$(call target)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go install -a -v $(strip $(GO_FLAGS)) $(CMD)


## test, bench and coverage

.PHONY: test
test: mod/vendor  ## Run the package test with checks race condition.
	$(call target)
	$(GO_TEST) -v -race $(strip $(GOFLAGS)) -run=$(GO_TEST_FUNC) $(GO_TEST_PKGS)

.PHONY: bench
bench: mod/vendor  ## Take a package benchmark.
	$(call target)
	$(GO_TEST) -v $(strip $(GOFLAGS)) -run='^$$' -bench=$(GO_BENCH_FUNC) -benchmem $(GO_TEST_PKGS)

.PHONY: bench/race
bench/race: mod/vendor  ## Take a package benchmark with checks race condition.
	$(call target)
	$(GO_TEST) -v -race $(strip $(GO_FLAGS)) -run='^$$' -bench=$(GO_BENCH_FUNC) -benchmem $(GO_TEST_PKGS)

.PHONY: bench/trace
bench/trace:  ## Take a package benchmark with take a trace profiling.
	$(GO_TEST) -v -c -o bench-trace.test $(PKG)/stackdriver
	GODEBUG=allocfreetrace=1 ./bench-trace.test -test.run=none -test.bench=$(GO_BENCH_FUNC) -test.benchmem -test.benchtime=10ms 2> trace.log

.PHONY: coverage
coverage: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
coverage: GO_LDFLAGS=${GO_LDFLAGS_STATIC}
coverage: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
coverage: clean  ## Take test coverage.
	$(call target)
	@$(GO_TEST) -v -race $(strip $(GO_FLAGS)) -covermode=atomic -coverpkg=$(PKG)/pkg/... -coverprofile=coverage.out $(GO_PKGS)

$(GO_PATH)/bin/go-junit-report:
	@GO111MODULE=off go get -u github.com/jstemmer/go-junit-report

.PHONY: cmd/go-junit-report
cmd/go-junit-report: $(GO_PATH)/bin/go-junit-report  # go get 'go-junit-report' binary

.PHONY: coverage/ci
coverage/ci: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
coverage/ci: GO_LDFLAGS=${GO_LDFLAGS_STATIC}
coverage/ci: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
coverage/ci: cmd/go-junit-report  ## Take test coverage.
	$(call target)
	@mkdir -p /tmp/ci/artifacts /tmp/ci/test-results
	$(GO_TEST) -v -race $(strip $(GO_FLAGS)) -covermode=atomic -coverpkg=$(PKG)/pkg/... -coverprofile=/tmp/ci/artifacts/coverage.out $(GO_PKGS) 2>&1 | tee /dev/stderr | go-junit-report -set-exit-code > /tmp/ci/test-results/junit.xml
	@go tool cover -html=/tmp/ci/artifacts/coverage.out -o /tmp/ci/artifacts/coverage.html


## lint

.PHONY: lint
lint: lint/golangci-lint  ## Run all linters.

.PHONY: $(GO_PATH)/bin/golangci-lint
$(GO_PATH)/bin/golangci-lint:
	@GO111MODULE=off go get -u -v github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: cmd/golangci-lint
cmd/golangci-lint: $(GO_PATH)/bin/golangci-lint  # go get 'golangci-lint' binary

.PHONY: golangci-lint
lint/golangci-lint: cmd/golangci-lint .golangci.yml mod/vendor  ## Run golangci-lint.
	$(call target)
	@GO111MODULE=on golangci-lint run ./...


## mod

.PHONY: mod/init
mod/init:
	$(call target)
	@GO111MODULE=on go mod init > /dev/null 2>&1 || true

.PHONY: mod/goget
mod/goget:  ## Update module and go.mod.
	$(call target)
	@GO111MODULE=on go get -u -m -v -x ./...

.PHONY: mod/tidy
mod/tidy:
	$(call target)
	@GO111MODULE=on go mod tidy -v

.PHONY: mod/vendor
mod/vendor:
	$(call target)
	@GO111MODULE=on go mod vendor

.PHONY: mod/graph
mod/graph:
	$(call target)
	@GO111MODULE=on go mod graph

.PHONY: mod/clean
mod/clean:
	$(call target)
	@$(RM) go.mod go.sum
	@$(RM) -r $(shell find vendor -maxdepth 1 -path "vendor/*" -type d)

.PHONY: mod/lock/go-client
mod/lock/go-client:  # locked to neovim/go-client@api/32405de
	$(call target)
	@go get -u -m -v -x github.com/neovim/go-client@api/32405de

.PHONY: mod/lock/delve
mod/lock/delve:  # locked to go-delve/delve@92dad94
	$(call target)
	@go mod edit -replace=github.com/googleapis/gax-go/v2@v2.0.0=github.com/googleapis/gax-go/v2@v2.0.3
	@go get -u -m -v -x github.com/derekparker/delve@92dad94
	@go get -u -m -v -x golang.org/x/arch@f4009597
	@go get -u -m -v -x golang.org/x/debug@fb50892
	@go get -u -m -v -x golang.org/x/sys@f3918c30c

.PHONY: mod/install
mod/install: GO_FLAGS+=-ldflags="${GO_LDFLAGS_STATIC}" -installsuffix netgo
mod/install: GO_BUILDTAGS+=netgo static static_build
mod/install: mod/lock/go-client mod/lock/delve mod/tidy mod/vendor
	$(call target)
	GO111MODULE=on go install -mod=vendor -v $(strip $(GO_FLAGS)) $(GO_VENDOR_PKGS) || @GO111MODULE=off go install -v $(strip $(GO_FLAGS)) $(GO_VENDOR_PKGS)

.PHONY: mod/update
mod/update: mod/goget mod/lock/go-client mod/lock/delve mod/tidy mod/vendor mod/install  ## Updates all vendor packages.
	@sed -i ':a;N;$$!ba;s|go 1\.12\n\n||g' go.mod

.PHONY: mod
mod: mod/tidy mod/vendor mod/install  ## Updates the vendoring directory via go mod.
	@sed -i ':a;N;$$!ba;s|go 1\.12\n\n||g' go.mod


## miscellaneous

.PHONY: container/build
container/build: Dockerfile mod/vendor  ## Build the zchee/nvim-go docker container for testing on the Linux.
	docker image build --rm --force-rm --pull --progress=plain -t $(IMAGE_REGISTRY)/$(APP):$(VERSION:v%=%) .

.PHONY: container/build/nocache
container/build/nocache:  ## Build the zchee/nvim-go docker container for testing on the Linux without cache.
	$(call target)
	docker image build --rm --force-rm --no-cache --pull --progress=plain -t -t $(IMAGE_REGISTRY)/$(APP):$(VERSION:v%=%) .

.PHONY: container/push
container/push:  ## Push the container image to $IMAGE_REGISTRY.
	docker image push $(IMAGE_REGISTRY)/$(APP):$(VERSION:v%=%)


.PHONY: boilerplate/go/%
boilerplate/go/%: BOILERPLATE_PKG_DIR=$(shell printf $@ | cut -d'/' -f3- | rev | cut -d'/' -f2- | rev)
boilerplate/go/%: BOILERPLATE_PKG_NAME=$(if $(findstring $@,cmd),main,$(shell printf $@ | rev | cut -d/ -f2 | rev))
boilerplate/go/%: hack/boilerplate/boilerplate.go.txt  ## Creates go file to % location based from boilerplate.go.txt.
	@if [ ! -d ${BOILERPLATE_PKG_DIR} ]; then mkdir -p ${BOILERPLATE_PKG_DIR}; fi
	@cat hack/boilerplate/boilerplate.go.txt <(printf "package ${BOILERPLATE_PKG_NAME}\\n") > $*
	@sed -i "s|YEAR|$(shell date '+%Y')|g" $*


.PHONY: clean
clean:  ## Cleanups the any build binaries or packages.
	$(call target)
	@$(RM) -r ./bin *.out *.test *.prof trace.log


.PHONY: AUTHORS
AUTHORS:  ## Creates AUTHORS file.
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@printf "$(shell git log --format="\n%aN <%aE>" | LC_ALL=C.UTF-8 sort -uf)" >> $@


.PHONY: todo
todo:  ## Print the all of (TODO|BUG|XXX|FIXME|NOTE) in nvim-go package sources
	@rg -e '(TODO|BUG|XXX|FIXME|NOTE)(\(.+\):|:)' --follow --hidden --glob='!.git' --glob='!vendor' --glob='!internal' --glob='!Makefile' --glob='!snippets' --glob='!indent'


.PHONY: help
help:  ## Show make target help.
	@perl -nle 'BEGIN {printf "Usage:\n  make \033[33m<target>\033[0m\n\nTargets:\n"} printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 if /^([a-zA-Z\/_-].+)+:.*?\s+## (.*)/' ${MAKEFILE_LIST}

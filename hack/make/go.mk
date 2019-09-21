# ----------------------------------------------------------------------------
# global

SHELL = /usr/bin/env bash

ifneq ($(shell command -v go),)
GO_PATH ?= $(shell go env GOPATH)
GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)

ifneq ($(wildcard vendor),)  # exist vendor directory
PKG := $(subst $(GO_PATH)/src/,,$(CURDIR))
GO_PKGS := $(shell go list ./... | grep -v -e 'cmd' -e '.pb.go')
GO_APP_PKGS := $(shell go list -f '{{if and (or .GoFiles .CgoFiles) (ne .Name "main")}}{{.ImportPath}}{{end}}' ${PKG}/...)
GO_TEST_PKGS := $(shell go list -f='{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)
GO_VENDOR_PKGS=
ifneq ($(wildcard ./vendor),)
	GO_VENDOR_PKGS = $(shell go list -f '{{if and (or .GoFiles .CgoFiles) (ne .Name "main")}}./vendor/{{.ImportPath}}{{end}}' ./vendor/...)
endif
endif

GO_TEST ?= go test
ifneq ($(shell command -v gotest),)
	GO_TEST=gotest
endif
GO_TEST_FUNC ?= .
GO_TEST_FLAGS ?=
GO_BENCH_FUNC ?= .
GO_BENCH_FLAGS ?= -benchmem

ifneq ($(shell command -v git),)
VERSION := $(shell cat VERSION.txt)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_UNTRACKED_CHANGES:= $(shell git status --porcelain --untracked-files=no)
ifneq ($(GIT_UNTRACKED_CHANGES),)
	GIT_COMMIT := $(GIT_COMMIT)-dirty
endif
CTIMEVAR:=-X ${PKG}/pkg/version.version=$(VERSION) -X ${PKG}/pkg/version.gitCommit=$(GIT_COMMIT)
endif

CGO_ENABLED ?= 0
GO_GCFLAGS=
GO_LDFLAGS=-s -w $(CTIMEVAR)
ifneq ($(GO_OS),darwin)
GO_LDFLAGS+='-extldflags=-static'
ifeq ($(CI),)
	GO_LDFLAGS+=-d
endif
endif

GO_BUILDTAGS=osusergo
ifneq ($(GO_OS),darwin)
	GO_BUILDTAGS+=netgo
endif
GO_BUILDTAGS_STATIC=static static_build
GO_INSTALLSUFFIX_STATIC=netgo
GO_FLAGS ?= -tags='$(GO_BUILDTAGS)' -gcflags="${GO_GCFLAGS}" -ldflags="${GO_LDFLAGS}"

GO_MOD_FLAGS =
ifneq ($(wildcard go.mod),)  # exist go.mod
ifneq ($(GO111MODULE),off)
	GO_MOD_FLAGS=-mod=vendor
endif
endif
endif

CONTAINER_REGISTRY := gcr.io/containerz
CONTAINER_BUILD_TAG ?= $(VERSION:v%=%)
CONTAINER_BUILD_ARGS_BASE = --rm --pull --build-arg GOLANG_VERSION=${GOLANG_VERSION} --build-arg ALPINE_VERSION=${ALPINE_VERSION}
ifneq (${SHORT_SHA},)  ## for Cloud Build
	CONTAINER_BUILD_ARGS_BASE+=--build-arg SHORT_SHA=${SHORT_SHA}
endif
CONTINUOUS_INTEGRATION ?=  ## for buildkit
ifneq (${CONTINUOUS_INTEGRATION},)
	CONTAINER_BUILD_ARGS_BASE+=--progress=plain
endif
CONTAINER_BUILD_ARGS ?= ${CONTAINER_BUILD_ARGS_BASE}
CONTAINER_BUILD_TARGET ?= ${APP}
ifneq (${CONTAINER_BUILD_TARGET},${APP})
	CONTAINER_BUILD_TAG=${VERSION}-${CONTAINER_BUILD_TARGET}
endif
CONTAINER_RUN_ARGS ?= --rm --interactive --tty

# ----------------------------------------------------------------------------
# defines

GOPHER = ""
define target
@printf "$(GOPHER)  \\x1b[1;32m$(patsubst ,$@,$(1))\\x1b[0m\\n"
endef

# ----------------------------------------------------------------------------
# targets

## build and install

.PHONY: bin/$(APP)
bin/$(APP): VERSION.txt
	$(call target)
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go build -v $(strip $(GO_FLAGS)) -o $@ $(CMD)

.PHONY: build
build: GO_FLAGS+=${GO_MOD_FLAGS}
build: bin/$(APP)  ## Builds a dynamic executable or package.

.PHONY: static
static: GO_FLAGS+=${GO_MOD_FLAGS}
static: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
static: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
static: bin/$(APP)  ## Builds a static executable or package.

.PHONY: install
install: GO_FLAGS+=${GO_MOD_FLAGS}
install: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
install: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
install:  ## Installs the executable or package.
	$(call target)
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go install -v $(strip $(GO_FLAGS)) $(CMD)

.PHONY: pkg/install
pkg/install: GO_FLAGS+=${GO_MOD_FLAGS}
pkg/install: GO_LDFLAGS=
pkg/install: GO_BUILDTAGS=
pkg/install:
	$(call target)
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GO_OS) GOARCH=$(GO_ARCH) go install -v ${GO_APP_PKGS}

## test, bench and coverage

.PHONY: test
test: CGO_ENABLED=1
test: GO_FLAGS+=${GO_MOD_FLAGS}
test: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
test: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
test:  ## Runs package test including race condition.
	$(call target)
	@GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(GO_TEST) -v -race $(strip $(GO_FLAGS)) -run=$(GO_TEST_FUNC) $(GO_TEST_PKGS)

.PHONY: bench
bench: GO_FLAGS+=${GO_MOD_FLAGS}
bench: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
bench: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
bench:  ## Take a package benchmark.
	$(call target)
	@GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(GO_TEST) -v $(strip $(GO_FLAGS)) -run='^$$' -bench=$(GO_BENCH_FUNC) -benchmem $(GO_TEST_PKGS)

.PHONY: bench/race
bench/race: CGO_ENABLED=1
bench/race:  ## Takes packages benchmarks with the race condition.
	$(call target)
	@GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(GO_TEST) -v -race $(strip $(GO_FLAGS)) -run='^$$' -bench=$(GO_BENCH_FUNC) -benchmem $(GO_TEST_PKGS)

.PHONY: bench/trace
bench/trace:  ## Take a package benchmark with take a trace profiling.
	$(GO_TEST) -v -c -o bench-trace.test $(PKG)
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) GODEBUG=allocfreetrace=1 ./bench-trace.test -test.run=none -test.bench=$(GO_BENCH_FUNC) -test.benchmem -test.benchtime=10ms 2> trace.log

.PHONY: coverage
coverage: CGO_ENABLED=1
coverage: GO_FLAGS+=${GO_MOD_FLAGS}
coverage: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
coverage: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
coverage:  ## Takes packages test coverage.
	$(call target)
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(GO_TEST) -v -race $(strip $(GO_FLAGS)) -covermode=atomic -coverpkg=$(PKG)/... -coverprofile=coverage.out $(GO_PKGS)

$(GO_PATH)/bin/go-junit-report:
	@GO111MODULE=off go get -u github.com/jstemmer/go-junit-report

.PHONY: cmd/go-junit-report
cmd/go-junit-report: $(GO_PATH)/bin/go-junit-report  # go get 'go-junit-report' binary

.PHONY: coverage/ci
coverage/ci: CGO_ENABLED=1
coverage/ci: GO_FLAGS+=-mod=readonly
coverage/ci: GO_BUILDTAGS+=${GO_BUILDTAGS_STATIC}
coverage/ci: GO_FLAGS+=-installsuffix ${GO_INSTALLSUFFIX_STATIC}
coverage/ci: cmd/go-junit-report
coverage/ci:  ## Takes packages test coverage, and output coverage results to CI artifacts.
	$(call target)
	@mkdir -p /tmp/ci/artifacts /tmp/ci/test-results
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(GO_TEST) -a -v -race $(strip $(GO_FLAGS)) -covermode=atomic -coverpkg=$(PKG)/... -coverprofile=/tmp/ci/artifacts/coverage.out $(GO_PKGS) 2>&1 | tee /dev/stderr | go-junit-report -set-exit-code > /tmp/ci/test-results/junit.xml
	@if [[ -f '/tmp/ci/artifacts/coverage.out' ]]; then go tool cover -html=/tmp/ci/artifacts/coverage.out -o /tmp/ci/artifacts/coverage.html; fi


## lint

.PHONY: lint
lint: lint/vet lint/golangci-lint  ## Run all linters.

$(GO_PATH)/bin/vet:
	@GO111MODULE=off go get -u golang.org/x/tools/go/analysis/cmd/vet golang.org/x/tools/go/analysis/passes/...

.PHONY: cmd/vet
cmd/vet: $(GO_PATH)/bin/vet  # go get 'vet' binary

.PHONY: lint/vet
lint/vet: cmd/vet
	$(call target)
	@GO111MODULE=on vet -asmdecl -assign -atomic -atomicalign -bools -buildtag -cgocall -compositewhitelist -copylocks -errorsas -httpresponse -loopclosure -lostcancel -methods -nilfunc -printfuncs -rangeloops -shift -source -stdmethods -structtag -tests -unmarshal -unreachable -unsafeptr $(GO_PKGS)

$(GO_PATH)/bin/golangci-lint:
	@GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: cmd/golangci-lint
cmd/golangci-lint: $(GO_PATH)/bin/golangci-lint  # go get 'golangci-lint' binary

.PHONY: lint/golangci-lint
lint/golangci-lint: cmd/golangci-lint .golangci.yml  ## Run golangci-lint.
	$(call target)
	@GO111MODULE=on GOGC=off golangci-lint run ./...


## mod

.PHONY: mod/init
mod/init:  ## Initializes and writes a new `go.mod` to the current directory.
	$(call target)
	@GO111MODULE=on go mod init > /dev/null 2>&1 || true

.PHONY: mod/get
mod/get:  ## Updates all module packages and go.mod.
	$(call target)
	@GO111MODULE=on go get -u -m -v -x

.PHONY: mod/tidy
mod/tidy:  ## Makes sure go.mod matches the source code in the module.
	$(call target)
	@GO111MODULE=on go mod tidy -v

.PHONY: mod/vendor
mod/vendor: mod/tidy  ## Resets the module's vendor directory and fetch all modules packages.
	$(call target)
	@GO111MODULE=on go mod vendor -v

.PHONY: mod/graph
mod/graph:  ## Prints the module requirement graph with replacements applied.
	$(call target)
	@GO111MODULE=on go mod graph

.PHONY: mod/install
mod/install: mod/tidy mod/vendor
mod/install:  ## Install the module vendor package as an object file.
	$(call target)
	@GO111MODULE=off go install -v $(strip $(GO_FLAGS)) $(GO_VENDOR_PKGS) || GO111MODULE=on go install -mod=vendor -v $(strip $(GO_FLAGS)) $(GO_VENDOR_PKGS)

.PHONY: mod/update
mod/update: mod/get mod/tidy mod/vendor mod/install  ## Updates all of vendor packages.

.PHONY: mod
mod: mod/init mod/tidy mod/vendor mod/install
mod:  ## Updates the vendoring directory using go mod.


## clean

.PHONY: clean
clean:  ## Cleanups binaries and extra files in the package.
	$(call target)
	@$(RM) -r ./bin *.out *.test *.prof trace.log


## container

.PHONY: container/build
container/build:  ## Creates the container image.
	docker image build ${CONTAINER_BUILD_ARGS} --target ${CONTAINER_BUILD_TARGET} -t $(CONTAINER_REGISTRY)/$(APP):${CONTAINER_BUILD_TAG} .

.PHONY: container/push
container/push:  ## Pushes the container image to $CONTAINER_REGISTRY.
	docker image push $(CONTAINER_REGISTRY)/$(APP):$(VERSION)


## boilerplate

.PHONY: boilerplate/go/%
boilerplate/go/%: BOILERPLATE_PKG_DIR=$(shell printf $@ | cut -d'/' -f3- | rev | cut -d'/' -f2- | rev)
boilerplate/go/%: BOILERPLATE_PKG_NAME=$(if $(findstring $@,cmd),main,$(shell printf $@ | rev | cut -d/ -f2 | rev))
boilerplate/go/%: hack/boilerplate/boilerplate.go.txt
boilerplate/go/%:  ## Creates a go file based on boilerplate.go.txt in % location.
	@if [ ! -d ${BOILERPLATE_PKG_DIR} ]; then mkdir -p ${BOILERPLATE_PKG_DIR}; fi
	@cat hack/boilerplate/boilerplate.go.txt <(printf "package ${BOILERPLATE_PKG_NAME}\\n") > $*
	@sed -i "s|YEAR|$(shell date '+%Y')|g" $*

## miscellaneous

.PHONY: AUTHORS
AUTHORS:  ## Creates AUTHORS file.
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@printf "$(shell git log --format="\n%aN <%aE>" | LC_ALL=C.UTF-8 sort -uf)" >> $@

.PHONY: TODO
TODO:  ## Print the all of (TODO|BUG|XXX|FIXME|NOTE) in packages.
	@rg -e '(TODO|BUG|XXX|FIXME|NOTE)(\(.+\):|:)' --follow --hidden --glob='!.git' --glob='!vendor' --glob='!internal' --glob='!Makefile' --glob='!snippets' --glob='!indent'

.PHONY: help
help:  ## Show make target help.
	@perl -nle 'BEGIN {printf "Usage:\n  make \033[33m<target>\033[0m\n\nTargets:\n"} printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 if /^([a-zA-Z\/_-].+)+:.*?\s+## (.*)/' ${MAKEFILE_LIST}

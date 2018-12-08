# ----------------------------------------------------------------------------
# global

# base
SHELL := /usr/bin/env bash
GO_PATH := $(shell go env GOPATH)
CGO_ENABLED ?= 1

# pkg
PKG := $(subst $(GO_PATH)/src/,,$(CURDIR))
GO_PKGS := $(shell go list ./... | grep -v -e '.pb.go')
GO_ABS_PKGS := $(shell go list -f '$(GO_PATH)/src/{{.ImportPath}}' ./... | grep -v -e '.pb.go')
GO_TEST_PKGS := $(shell go list -f='{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)

# version
VERSION=$(shell cat VERSION.txt || devel)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_UNTRACKED_CHANGES:= $(shell git status --porcelain --untracked-files=no)
ifneq ($(GIT_UNTRACKED_CHANGES),)
	GIT_COMMIT := $(GIT_COMMIT)-dirty
endif
# CTIMEVAR=-X=$(PKG)/main.tag=$(VERSION) -X=$(PKG)/main.gitCommit=$(GIT_COMMIT)
CTIMEVAR=-X $(PKG)/pkg/version.Tag=$(VERSION) -X $(PKG)/pkg/version.GitCommit=$(GIT_COMMIT)
# CTIMEVAR=-X $(PKG)/pkg/version.Version=$(VERSION)@$(GIT_COMMIT)

# gcflags, ldflags
GO_GCFLAGS?=
GO_LDFLAGS=$(CTIMEVAR)
GO_LDFLAGS_STATIC=$(CTIMEVAR) '-extldflags=-static'
# GO_LDFLAGS_STATIC=${GO_LDFLAGS}+='-extldflags=-static'
ifneq ($(NVIM_GO_DEBUG),)
	GO_LDFLAGS+=-w -s
else
	GO_GCFLAGS+=all='-N -l -dwarflocationlists=true'  # https://tip.golang.org/doc/diagnostics.html#debugging
	GO_LDFLAGS+=-compressdwarf=false
endif

# build tags
GO_BUILD_TAGS ?= osusergo netgo

# GOFLAGS
GOFLAGS ?= -tags '$(GO_BUILD_TAGS)' -installsuffix netgo
# ifneq ($(CI),)
# 	GOFLAGS+=-mod=vendor
# endif

ifneq ($(GO_GCFLAGS),)
	GOFLAGS+=-gcflags $(strip $(GO_GCFLAGS))
endif
ifneq ($(GO_LDFLAGS),)
	GOFLAGS+=-ldflags '$(strip $(GO_LDFLAGS))'
endif

# test
GO_TEST ?= go test
GO_TEST_FUNC ?= .
GO_BENCH_FUNC ?= .

# lint
GOLANGCI_EXCLUDE ?=
ifeq ($(wildcard '.errcheckignore'),)
	GOLANGCI_EXCLUDE=$(foreach pat,$(shell cat .errcheckignore),--exclude '$(pat)')
endif
GOLANGCI_CONFIG ?=
ifeq ($(wildcard '.golangci.yml'),)
	GOLANGCI_CONFIG+=--config .golangci.yml
endif

# docker
IMAGE_REGISTRY := quay.io/zchee

# ----------------------------------------------------------------------------
# defines

define target
@printf "+ \\033[32m$(patsubst ,$@,$(1))\\033[0m\\n"
endef

# ----------------------------------------------------------------------------
# targets

.PHONY: bin/$(APP)
bin/$(APP):
	CGO_ENABLED=$(CGO_ENABLED) go build -v -o ./bin/${APP} $(strip $(GOFLAGS)) ./cmd/${APP}

.PHONY: $(APP)
$(APP): bin/$(APP)

.PHONY: build
build: $(APP)  ## Builds a dynamic executable or package.
	$(call target)

.PHONY: build/race
build/race: GOFLAGS+=-race
build/race: clean $(APP)  ## Build the nvim-go binary with race
	$(call target)

.PHONY: static
static: GOFLAGS+=${GO_LDFLAGS_STATIC}
static: $(APP)  ## Builds a static executable or package.
	$(call target)


## test, bench and coverage

.PHONY: test
test:  ## Run the package test with checks race condition.
	$(call target)
	$(GO_TEST) -v -race $(strip $(GOFLAGS)) -run=$(GO_TEST_FUNC) $(GO_TEST_PKGS)

# .PHONY: test/cpu
# test/cpu: GOFLAGS+=-cpuprofile cpu.out
# test/cpu: _test  ## Run the package test with take a cpu profile.
# 	$(call target)
#
# .PHONY: test/mem
# test/mem: GOFLAGS+=-memprofile mem.out
# test/mem: _test  ## Run the package test with take a memory profile.
# 	$(call target)
#
# .PHONY: test/mutex
# test/mutex: GOFLAGS+=-mutexprofile mutex.out
# test/mutex: _test  ## Run the package test with take a mutex profile.
# 	$(call target)
#
# .PHONY: test/block
# test/block: GOFLAGS+=-blockprofile block.out
# test/block: _test  ## Run the package test with take a blockingh profile.
# 	$(call target)
#
# .PHONY: test/trace
# test/trace: GOFLAGS+=-trace trace.out
# test/trace: _test  ## Run the package test with take a trace profile.
# 	$(call target)

.PHONY: bench
bench:  ## Take a package benchmark.
	$(call target)
	$(GO_TEST) -v $(strip $(GOFLAGS)) -run='^$$' -bench=$(GO_BENCH_FUNC) -benchmem $(GO_TEST_PKGS)

# .PHONY: bench/cpu
# bench/cpu: GOFLAGS+=-cpuprofile cpu.out
# bench/cpu: bench  ## Take a package benchmark with take a cpu profile.
#
# .PHONY: bench/trace
# bench/trace:  ## Take a package benchmark with take a trace profile.
# 	$(call target)
# 	$(GO_TEST) -v -c -o bench-trace.test $(PKG)/stackdriver
# 	GODEBUG=allocfreetrace=1 ./bench-trace.test -test.run=none -test.bench=$(GO_BENCH_FUNC) -test.benchmem -test.benchtime=10ms 2> trace.log

.PHONY: coverage
coverage:  ## Take test coverage.
	$(call target)
	$(GO_TEST) -v -race $(strip $(GOFLAGS)) -covermode=atomic -coverpkg=$(PKG)/... -coverprofile=coverage.out $(GO_TEST_PKGS)

$(GO_PATH)/bin/go-junit-report:
	@GO111MODULE=off go get -u github.com/jstemmer/go-junit-report

cmd/go-junit-report: $(GO_PATH)/bin/go-junit-report  ## go get `go-junit-report` binary.

.PHONY: coverage/junit
coverage/junit: cmd/go-junit-report  ## Take test coverage and output test results with junit syntax.
	$(call target)
	mkdir -p _test-results
	$(GO_TEST) -v -race $(strip $(GOFLAGS)) -covermode=atomic -coverpkg=$(PKG)/... -coverprofile=coverage.out $(GO_TEST_PKGS) 2>&1 | tee /dev/stderr | go-junit-report -set-exit-code > _test-results/report.xml


## lint

lint: lint/fmt lint/govet lint/golint lint/golangci-lint  ## Run all linters.

.PHONY: lint/fmt
lint/fmt:  ## Verifies all files have been `gofmt`ed.
	$(call target)
	@gofmt -s -l . | grep -v -e 'vendor' -e '.pb.go' | tee /dev/stderr

.PHONY: lint/govet
lint/govet:  ## Verifies `go vet` passes.
	$(call target)
	@go vet -all $(GO_PKGS) | tee /dev/stderr

$(GO_PATH)/bin/golint:
	@GO111MODULE=off go get -u golang.org/x/lint/golint

cmd/golint: $(GO_PATH)/bin/golint  ## go get `golint` binary.

.PHONY: lint/golint
lint/golint: cmd/golint  ## Verifies `golint` passes.
	$(call target)
	@golint -min_confidence=0.3 $(GO_PKGS)

$(GO_PATH)/bin/golangci-lint:
	@GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

cmd/golangci-lint: $(GO_PATH)/bin/golangci-lint  ## Install `golangci-lint` command.

.PHONY: golangci-lint
lint/golangci-lint: cmd/golangci-lint  ## Run golangci-lint.
	$(call target)
	@golangci-lint run $(strip $(GOLANGCI_CONFIG)) ./...


## vendor

### dep

$(GO_PATH)/bin/dep:
	@go get -u github.com/golang/dep/cmd/dep

cmd/dep: $(GO_PATH)/bin/dep  ## go get `dep` binary.

.PHONY: dep/init
dep/init: cmd/dep  ## Init dep files.
	$(call target)
	@dep init -v -no-examples

.PHONY: dep/ensure
dep/ensure: cmd/dep Gopkg.toml  ## Fetchs the vendor packages via dep ensure.
	$(call target)
	@dep ensure -v

.PHONY: dep/ensure/only-vendor
dep/ensure/vendor: cmd/dep Gopkg.toml Gopkg.lock  ## Fetchs the vendor packages only via dep ensure.
	$(call target)
	@dep ensure -v -vendor-only

.PHONY: dep/update
dep/update: cmd/dep  ## Updates the vendor packages via dep.
	$(call target)
	@dep ensure -v -update

.PHONY: dep/clean
dep/clean: cmd/dep  ## Cleanups the dep vendoring tool files.
	$(call target)
	@$(RM) Gopkg.toml Gopkg.lock

.PHONY: dep
dep: dep/ensure dep/update  ## Updates all vendor packages via dep.

### mod

.PHONY: mod/init
mod/init:  ## Init go.mod file.
	$(call target)
	@GO111MODULE=on go mod init

.PHONY: mod/tidy
mod/tidy:  ## Makes sure go.mod matches the source code in the module.
	$(call target)
	@GO111MODULE=on go mod tidy -v

.PHONY: mod/vendor
mod/vendor: go.mod go.sum  ## Fetchs the vendor packages via go mod.
	$(call target)
	@GO111MODULE=on go mod vendor -v

.PHONY: mod/graph
mod/graph:  ## Prints the module requirement graph with replacements applied in text form.
	$(call target)
	@GO111MODULE=on go mod graph

.PHONY: mod/clean
mod/clean:  ## Cleanups the go mod vendoring tool files.
	$(call target)
	@$(RM) go.mod go.sum

.PHONY: mod
mod: Gopkg.toml Gopkg.lock mod/clean mod/init mod/tidy  ## Updates the vendor packages via go mod.
	@sed -i ':a;N;$$!ba;s|go 1\.12\n\n||g' go.mod

.PHONY: vendor
vendor: dep mod  ## Updates all vendor packages.


## miscellaneous

boilerplate/go/%: BOILERPLATE_PKG_DIR=$(shell printf $@ | cut -d'/' -f3- | rev | cut -d'/' -f2- | rev)
boilerplate/go/%: BOILERPLATE_PKG_NAME=$(if $(findstring $@,cmd),main,$(shell printf $@ | rev | cut -d/ -f2 | rev))
boilerplate/go/%: hack/boilerplate/boilerplate.go.txt  ## Create initial .go file from the boilerplate.go.txt
	@if [ ! -d ${BOILERPLATE_PKG_DIR} ]; then mkdir -p ${BOILERPLATE_PKG_DIR}; fi
	@cat hack/boilerplate/boilerplate.go.txt <(printf "package ${BOILERPLATE_PKG_NAME}\\n") > $*


.PHONY: AUTHORS
AUTHORS:  ## Creates AUTHORS file.
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@printf "$(shell git log --format="\n%aN <%aE>" | LC_ALL=C.UTF-8 sort -uf)" >> $@


.PHONY: clean
clean:  ## Cleanups the any build binaries or packages.
	$(call target)
	$(RM) -r ./bin *.out *.test *.prof trace.log


.PHONY: todo
todo:  ## Print the all of (TODO|BUG|XXX|FIXME|NOTE) in nvim-go package sources
	@pt -e '(TODO|BUG|XXX|FIXME|NOTE)(\(.+\):|:)' --follow --hidden --ignore=.git --ignore=vendor --ignore=internal --ignore=Makefile --ignore=snippets --ignore=indent


.PHONY: help
help:  ## Show make target help.
	@perl -nle 'BEGIN {printf "Usage:\n  make \033[33m<target>\033[0m\n\nTargets:\n"} printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 if /^([a-zA-Z\/_-].+)+:.*?\s+## (.*)/' ${MAKEFILE_LIST}

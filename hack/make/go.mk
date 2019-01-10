# ----------------------------------------------------------------------------
# global

# base
SHELL := /usr/bin/env bash
GO_PATH := $(shell go env GOPATH)

# pkg
PKG := $(subst $(GO_PATH)/src/,,$(CURDIR))
GO_PKGS := $(shell go list ./... | grep -v -e '.pb.go')
GO_ABS_PKGS := $(shell go list -f '$(GO_PATH)/src/{{.ImportPath}}' ./... | grep -v -e '.pb.go')
GO_TEST_PKGS := $(shell go list -f='{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)
GO_VENDOR_PKGS := $(shell go list -f '{{if and (or .GoFiles .CgoFiles) (ne .Name "main")}}./vendor/{{.ImportPath}}{{end}}' ./vendor/...)

# version
VERSION=$(shell cat VERSION.txt)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_UNTRACKED_CHANGES:= $(shell git status --porcelain --untracked-files=no)
ifneq ($(GIT_UNTRACKED_CHANGES),)
	GIT_COMMIT := $(GIT_COMMIT)-dirty
endif
CTIMEVAR=-X $(PKG)/pkg/version.Tag=$(VERSION) -X $(PKG)/pkg/version.GitCommit=$(GIT_COMMIT)

CGO_ENABLED ?= 0
GO_BUILD_TAGS := osusergo netgo
GO_FLAGS ?= -tags '$(GO_BUILD_TAGS)'

ifeq ($(NVIM_GO_DEBUG),)
	GO_LDFLAGS:=-ldflags "-w -s $(CTIMEVAR)"
	GO_LDFLAGS_STATIC:=-ldflags "-w -s $(CTIMEVAR) -extldflags -static"
else
	GO_GCFLAGS:=-gcflags all='-N -l -dwarflocationlists=true'  # https://tip.golang.org/doc/diagnostics.html#debugging
	GO_LDFLAGS:=-ldflags "compressdwarf=false $(CTIMEVAR)"
endif

ifneq ($(wildcard go.mod),)  # exist go.mod file
ifeq ($(CI),)  # $CI is empty
	GOFLAGS+=-mod=vendor
endif
endif

GO_TEST ?= go test
GO_TEST_FUNC ?= .
GO_TEST_FLAGS ?=
GO_BENCH_FUNC ?= .
GO_BENCH_FLAGS ?= -benchmem

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
bin/$(APP): mod/vendor
	$(call target,$@)
	CGO_ENABLED=$(CGO_ENABLED) go build -v $(strip $(GOFLAGS)) -o $@ $(CMD)

.PHONY: $(APP)
$(APP): bin/$(APP)

.PHONY: build
build: GO_FLAGS+=$(GO_GCFLAGS) ${GO_LDFLAGS}
build: $(APP)  ## Builds a dynamic executable or package.

.PHONY: build/race
build/race: GO_FLAGS+=-race
build/race: GO_FLAGS+=$(GO_GCFLAGS) ${GO_LDFLAGS}
build/race: clean $(APP) mod/vendor  ## Build the nvim-go binary with race
	$(call target)

.PHONY: static
static: GO_BUILD_TAGS+=static
static: GO_FLAGS+=$(GO_GCFLAGS) ${GO_LDFLAGS_STATIC}
static: $(APP) mod/vendor  ## Builds a static executable or package.
	$(call target)

.PHONY: install
install: GO_FLAGS+=${GO_LDFLAGS_STATIC}
install: mod/vendor  ## Installs the executable or package.
	$(call target)
	CGO_ENABLED=$(CGO_ENABLED) go install -a -v $(strip $(GO_FLAGS)) $(CMD)


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
coverage: mod/vendor  ## Take test coverage.
	$(call target)
	$(GO_TEST) -v -race $(strip $(GOFLAGS)) -covermode=atomic -coverpkg=$(PKG)/... -coverprofile=coverage.out $(GO_TEST_PKGS)

.PHONY: $(GO_PATH)/bin/go-junit-report
$(GO_PATH)/bin/go-junit-report:
	@GO111MODULE=off go get -u github.com/jstemmer/go-junit-report

.PHONY: cmd/go-junit-report
cmd/go-junit-report: $(GO_PATH)/bin/go-junit-report  # go get 'go-junit-report' binary

.PHONY: coverage/junit
coverage/junit: cmd/go-junit-report mod/vendor  ## Take test coverage and output test results with junit syntax.
	$(call target)
	@mkdir -p test-results
	$(GO_TEST) -v -race $(strip $(GO_FLAGS)) -covermode=atomic -coverpkg=$(PKG)/... -coverprofile=coverage.out $(GO_PKGS) 2>&1 | tee /dev/stderr | go-junit-report -set-exit-code > test-results/report.xml


## lint

.PHONY: lint
lint: lint/fmt lint/golangci-lint  ## Run all linters.

.PHONY: lint/fmt
lint/fmt:  ## Verifies all files have been `gofmt`ed.
	$(call target)
	@gofmt -s -l . 2>&1 | grep -v -E -e 'testdata' -e 'vendor' -e '\.pb.go' -e '_.*' | tee /dev/stderr

.PHONY: $(GO_PATH)/bin/golangci-lint
$(GO_PATH)/bin/golangci-lint:
	@GO111MODULE=off go get -u -v github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: cmd/golangci-lint
cmd/golangci-lint: $(GO_PATH)/bin/golangci-lint  # go get 'golangci-lint' binary

.PHONY: golangci-lint
lint/golangci-lint: cmd/golangci-lint .golangci.yml mod/vendor  ## Run golangci-lint.
	$(call target)
	@golangci-lint run ./...


## mod

.PHONY: mod/init
mod/init:
	$(call target)
	@GO111MODULE=on go mod init

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
	@$(RM) -r vendor

.PHONY: mod/lock/go-client
mod/lock/go-client:  # locked to neovim/go-client@api/32405de
	$(call target)
	@go get -u -m -v -x github.com/neovim/go-client@api/32405de

.PHONY: mod/lock/delve
mod/lock/delve:  # locked to derekparker/delve@92dad94
	$(call target)
	@go get -u -m -v -x github.com/derekparker/delve@92dad94 golang.org/x/arch@f40095975f golang.org/x/debug@fb508927b4 golang.org/x/sys@f3918c30c5

.PHONY: mod
mod: mod/clean mod/init mod/lock/go-client mod/lock/delve mod/tidy mod/vendor  ## Updates the vendoring directory via go mod.
	@sed -i ':a;N;$$!ba;s|go 1\.12\n\n||g' go.mod

.PHONY: mod/install
mod/install: mod/lock/go-client mod/lock/delve mod/tidy mod/vendor
	$(call target)
	@GO111MODULE=off go install -v $(GO_VENDOR_PKGS) || GO111MODULE=on go install -mod=vendor -v $(GO_VENDOR_PKGS)

.PHONY: mod/update
mod/update: mod/goget mod/lock/go-client mod/lock/delve mod/tidy mod/vendor mod/install  ## Updates all vendor packages.


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

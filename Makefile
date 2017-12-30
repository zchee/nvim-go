.DEFAULT_GOAL := build

# ----------------------------------------------------------------------------
# package level setting

APP := $(notdir $(CURDIR))
PACKAGE_ROOT := $(CURDIR)
PACKAGES := $(shell gb list ./...)

# ----------------------------------------------------------------------------
# common environment variables

SHELL := /usr/bin/env bash
CC := clang  # need compile cgo for delve
CXX := clang++
GO_LDFLAGS ?=
GO_GCFLAGS ?=
CGO_CFLAGS ?=
CGO_CPPFLAGS ?=
CGO_CXXFLAGS ?=
CGO_LDFLAGS ?=

# ----------------------------------------------------------------------------
# build and test flags

GB_PROJECT_DIR := $(shell gb env GB_PROJECT_DIR)
INTERNAL_GOPATH := ${GB_PROJECT_DIR}:${GB_PROJECT_DIR}/vendor
GO_BUILD_FLAGS ?=
GO_TEST_PKGS := $(shell gb list -f='{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./... | perl -pe 's/^\n//g')
GO_TEST_FUNCS ?= .
GO_TEST_FLAGS ?= -race -run=$(GO_TEST_FUNCS)
GO_BENCH_FUNCS ?= .
GO_BENCH_FLAGS ?= -run=^$$ -bench=${GO_BENCH_FUNCS} -benchmem

ifneq ($(NVIM_GO_DEBUG),)
GO_GCFLAGS+=-gcflags "-N -l -dwarflocationlists=true"  # https://tip.golang.org/doc/diagnostics.html#debugging
else
GO_LDFLAGS+=-ldflags "-w -s"
endif

ifneq ($(NVIM_GO_RACE),)
GO_BUILD_FLAGS+=-race
build: clean std-build-race
rebuild: std-build-race
manifest: APP=${APP}-race
endif

# ----------------------------------------------------------------------------
# targets

init:  # Install dependency tools
	go get -u -v \
		github.com/golang/lint/golint \
		honnef.co/go/tools/cmd/staticcheck \
		honnef.co/go/tools/cmd/gosimple \
		honnef.co/go/tools/cmd/errcheck-ng


build:  ## Build the nvim-go binary
	gb build ${GO_BUILD_FLAGS} ${GO_GCFLAGS} ${GO_LDFLAGS} ./cmd/...
.PHONY: build

rebuild: GO_BUILD_FLAGS+=-f
rebuild: clean build  ## Rebuild the nvim-go binary
.PHONY: rebuild

$(shell go env GOROOT)/pkg/darwin_amd64_race:
	go install -v -x -race std

std-build-race: $(shell go env GOROOT)/pkg/darwin_amd64_race  ## Build the Go stdlib runtime with -race

${CURDIR}/plugin/manifest: ${CURDIR}/plugin/manifest.go  ## Build the automatic writing neovim manifest utility binary
	go build -o ${CURDIR}/plugin/manifest ${CURDIR}/plugin/manifest.go

manifest: build ${CURDIR}/plugin/manifest  ## Write plugin manifest for developer
	${CURDIR}/plugin/manifest -w ${APP}
.PHONY: manifest

manifest-dump: build ${CURDIR}/plugin/manifest  ## Dump plugin manifest
	${CURDIR}/plugin/manifest -manifest ${APP}
.PHONY: manifest-dump


test: std-build-race  ## Run the package test
	gb test -v ${GO_TEST_FLAGS} ${GO_TEST_PKGS}
.PHONY: test

test-bench: GO_TEST_FLAGS+=${GO_BENCH_FLAGS}
test-bench: test ## Run the package test
.PHONY: test-bench


golint:
	@GOPATH=${INTERNAL_GOPATH} golint ${PACKAGES}
.PHONY: golint

errcheck-ng:
	@GOPATH=${INTERNAL_GOPATH} errcheck-ng ${PACKAGES}
.PHONY: errcheck-ng

gosimple:
	@GOPATH=${INTERNAL_GOPATH} gosimple ${PACKAGES}
.PHONY: gosimple

interfacer:
	@GOPATH=${INTERNAL_GOPATH} interfacer $(PACKAGES)
.PHONY: interfacer

staticcheck:
	@GOPATH=${INTERNAL_GOPATH} staticcheck $(PACKAGES)
.PHONY: staticcheck

unparam:
	@GOPATH=${INTERNAL_GOPATH} unparam $(PACKAGES)
.PHONY: unparam

vet:
	go vet ${PACKAGES}
.PHONY: vet

lint: golint errcheck-ng gosimple interfacer staticcheck unparam vet
.PHONY: lint


vendor-install:  # Install vendor packages for gocode completion
	go install -v -x ./vendor/...
.PHONY: vendor-install

vendor-update:  ## Update the all vendor packages
	gb vendor -d update -all
	${MAKE} vendor-clean
.PHONY: vendor-update

vendor-guru: vendor-guru-update vendor-guru-rename
.PHONY: vendor-guru

vendor-guru-update:  ## Update the internal guru package
	${RM} -r $(shell find ${PACKAGE_ROOT}/src/internal/guru -maxdepth 1 -type f -name '*.go' -not -name 'result.go')
	cp ${PACKAGE_ROOT}/vendor/golang.org/x/tools/cmd/guru/*.go ${PACKAGE_ROOT}/src/internal/guru
	sed -i "s|\t// TODO(adonovan): opt: parallelize.|\tbp.GoFiles = append(bp.GoFiles, bp.CgoFiles...)\n\n\0|" src/internal/guru/definition.go
	# ${RM} -r ${PACKAGE_ROOT}/src/internal/guru/guru_test.go ${PACKAGE_ROOT}/src/internal/guru/unit_test.go
.PHONY: vendor-guru-update

vendor-guru-rename: vendor-guru-update
	# Rename main to guru
	grep "package main" ${PACKAGE_ROOT}/src/internal/guru/*.go -l | xargs sed -i 's/package main/package guru/'
	# Add Result interface
	sed -i "s|PrintPlain(printf printfFunc)|\0\n\n\tResult(fset *token.FileSet) interface{}|" ${PACKAGE_ROOT}/src/internal/guru/guru.go
	# Export functions
	grep "findPackageMember" ${PACKAGE_ROOT}/src/internal/guru/*.go -l | xargs sed -i 's/findPackageMember/FindPackageMember/'
	grep "packageForQualIdent" ${PACKAGE_ROOT}/src/internal/guru/*.go -l | xargs sed -i 's/packageForQualIdent/PackageForQualIdent/'
	grep "guessImportPath" ${PACKAGE_ROOT}/src/internal/guru/*.go -l | xargs sed -i 's/guessImportPath/GuessImportPath/'
	# ignore build main.go
	sed -i "s|package guru // import \"golang.org/x/tools/cmd/guru\"|\n// +build ignore\n\n\0|" ${PACKAGE_ROOT}/src/internal/guru/main.go
	# ignore build guru_test.go
	sed -i "s|package guru_test|// +build ignore\n\n\0|" ${PACKAGE_ROOT}/src/internal/guru/guru_test.go
.PHONY: vendor-guru-rename

vendor-clean:  ## Cleanup vendor packages "*_test" files, testdata and nogo files.
	gb vendor -d purge
	@find ./vendor -type d -name 'testdata' -print | xargs rm -rf
	@find ./vendor -type f -name '*_test.go' -print -exec rm {} ";"
	@find ./vendor \
		\( -name '*.sh' \
		-or -name 'Makefile' \
		-or -name '*.yml' \
		-or -name '*.txtr' \
		-or -name '*.vim' \
		-or -name '*.el' \) \
		-type f -print -exec rm {} ";"
.PHONY: vendor-clean


clean:  ## Clean the {bin,pkg} directory
	${RM} -r ./bin ./pkg coverage.txt
.PHONY: clean


docker: docker-test  ## Run the docker container test on Linux
.PHONY: docker

docker-build:  ## Build the zchee/nvim-go docker container for testing on the Linux
	docker build --rm -t ${USER}/${APP} .
.PHONY: docker-build

docker-build-nocache:  ## Build the zchee/nvim-go docker container for testing on the Linux without cache
	docker build --rm --no-cache -t ${USER}/${APP} .
.PHONY: docker-build-nocache

docker-test: docker-build  ## Run the package test with docker container
	docker run --rm -it ${USER}/${APP} gb test -v ${GO_TEST_FLAGS} ${PACKAGES}
.PHONY: docker-test


todo:  ## Print the all of (TODO|BUG|XXX|FIXME|NOTE) in nvim-go package sources
	@pt -e 'TODO(\(.+\):|:)'  --after=1 --ignore vendor --ignore internal --ignore Makefile || true
	@pt -e 'BUG(\(.+\):|:)'   --after=1 --ignore vendor --ignore internal --ignore Makefile || true
	@pt -e 'XXX(\(.+\):|:)'   --after=1 --ignore vendor --ignore internal --ignore Makefile || true
	@pt -e 'FIXME(\(.+\):|:)' --after=1 --ignore vendor --ignore internal --ignore Makefile || true
	@pt -e 'NOTE(\(.+\):|:)'  --after=1 --ignore vendor --ignore internal --ignore Makefile || true
.PHONY: todo

help:  ## Print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.PHONY: help

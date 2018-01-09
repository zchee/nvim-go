.DEFAULT_GOAL := build

# ----------------------------------------------------------------------------
# package level setting

APP := $(notdir $(CURDIR))
PACKAGE_ROOT := $(CURDIR)
PACKAGES := $(shell go list ./src/...)
VENDOR_PACKAGES := $(shell go list -f='{{if and (or .GoFiles .CgoFiles) (ne .Name "main")}}{{.ImportPath}}{{end}}' ./vendor/...)

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

GO_TEST ?= go test
GO_BUILD_FLAGS ?=
GO_TEST_PKGS := $(shell go list -f='{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)
GO_TEST_FUNCS ?= .
GO_TEST_FLAGS ?= -race -run=$(GO_TEST_FUNCS)
GO_BENCH_FUNCS ?= .
GO_BENCH_FLAGS ?= -run=^$$ -bench=${GO_BENCH_FUNCS} -benchmem

ifneq ($(NVIM_GO_DEBUG),)
GO_GCFLAGS+=-gcflags=all="-N -l -dwarflocationlists=true"  # https://tip.golang.org/doc/diagnostics.html#debugging
else
GO_LDFLAGS+=-ldflags "-w -s"
endif

ifneq ($(NVIM_GO_RACE),)
GO_BUILD_FLAGS+=-race
build: clean
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
.PHONY: init

build:  ## Build the nvim-go binary
	go build -o ./bin/nvim-go -v ${GO_BUILD_FLAGS} ${GO_GCFLAGS} ${GO_LDFLAGS} ./cmd/nvim-go
.PHONY: build

rebuild: clean build  ## Rebuild the nvim-go binary
.PHONY: rebuild

manifest: build  ## Write plugin manifest for developer
	./bin/${APP} -manifest ${APP} -location ./plugin/${APP}.vim > /dev/null 2>&1
.PHONY: manifest

manifest-dump: build  ## Dump plugin manifest
	./bin/${APP} -manifest ${APP} 2>/dev/null
.PHONY: manifest-dump


test:  ## Run the package test
	${GO_TEST} -v ${GO_TEST_FLAGS} ${GO_TEST_PKGS}
.PHONY: test

bench: GO_TEST_FLAGS+=${GO_BENCH_FLAGS}
bench: test ## Take the packages benchmark
.PHONY: bench


golint:
	golint -set_exit_status -min_confidence=0.3 ${PACKAGES}
.PHONY: golint

errcheck-ng:
	errcheck-ng ${PACKAGES}
.PHONY: errcheck-ng

gosimple:
	gosimple ${PACKAGES}
.PHONY: gosimple

interfacer:
	interfacer $(PACKAGES)
.PHONY: interfacer

staticcheck:
	staticcheck $(PACKAGES)
.PHONY: staticcheck

unparam:
	unparam $(PACKAGES)
.PHONY: unparam

vet:
	go vet ${PACKAGES}
.PHONY: vet

lint: golint errcheck-ng gosimple interfacer staticcheck unparam vet
.PHONY: lint


vendor-install:  # Install vendor packages for gocode completion
	go install -v -x ${VENDOR_PACKAGES}
.PHONY: vendor-install

vendor-update:  ## Update the all vendor packages
	dep ensure -v -update
	dep prune -v
.PHONY: vendor-update

vendor-clean:  ## Cleanup unused files in vendor directory
	dep prune -v
.PHONY: vendor-clean

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


clean:  ## Clean the {bin,pkg} directory
	${RM} -r ./bin *.out *.prof
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
	docker run --rm -it ${USER}/${APP} go test -v ${GO_TEST_FLAGS} ${GO_TEST_PKGS}
.PHONY: docker-test


todo:  ## Print the all of (TODO|BUG|XXX|FIXME|NOTE) in nvim-go package sources
	@pt -e '(TODO|BUG|XXX|FIXME|NOTE)(\(.+\):|:)' --follow --hidden --ignore=.git --ignore=vendor --ignore=internal --ignore=Makefile
.PHONY: todo

help:  ## Print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
.PHONY: help

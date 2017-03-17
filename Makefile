GITHUB_USER := zchee

PACKAGE_NAME := nvim-go
PACKAGE_DIR := $(shell pwd)
BINARY_NAME := bin/nvim

CC := clang
CXX := clang++
GOPATH ?= $(shell go env GOPATH)
GOROOT ?= $(shell go env GOROOT)

GO_CMD := go
GB_CMD := $(GOPATH)/bin/gb
VENDOR_CMD := ${GB_CMD} vendor
DOCKER_CMD := docker

GO_LDFLAGS ?=
GO_GCFLAGS ?=
CGO_CFLAGS ?=
CGO_CPPFLAGS ?=
CGO_CXXFLAGS ?=
CGO_LDFLAGS ?=

GO_TEST_FLAGS ?= -v -race
test-bench: GO_TEST_FLAGS += -bench=. -benchmem

GO_BUILD := ${GB_CMD} build
GO_TEST := ${GB_CMD} test
GO_LINT := golint

ifneq ($(NVIM_GO_DEBUG),)
GO_GCFLAGS += -gcflags "-N -l"
else
GO_LDFLAGS += -ldflags "-w -s"
endif

ifneq ($(NVIM_GO_RACE),)
build: std-build-race
rebuild: std-build-race
GO_BUILD += -race
PACKAGE_NAME = nvim-go-race
endif

default: build

build:  ## Build the nvim-go binary
	${GO_BUILD} $(GO_LDFLAGS) ${GO_GCFLAGS}

rebuild: clean  ## Rebuild the nvim-go binary
	${GO_BUILD} -f $(GO_LDFLAGS) ${GO_GCFLAGS}

std-build-race:  ## Build the Go stdlib runtime with -race
	$(GO_CMD) install -v -x -race std

$(PACKAGE_DIR)/plugin/manifest:  ## Build the automatic writing neovim manifest utility binary
	${GO_BUILD} -o $(PACKAGE_DIR)/plugin/manifest $(PACKAGE_DIR)/plugin/manifest.go

manifest: build $(PACKAGE_DIR)/plugin/manifest  ## Write plugin manifest (for developers)
	$(PACKAGE_DIR)/plugin/manifest -w $(PACKAGE_NAME)

test: std-build-race  ## Run the package test
	${GO_TEST} $(GO_TEST_FLAGS)

test-docker: docker-run  ## Run the package test with docker container

test-bench:  ## Run the package test with -bench=.
	${GO_TEST} $(GO_TEST_FLAGS)

test-run:  ## Run the package test only those tests and examples
	${GO_TEST_RUN}


vendor-all:  ## Update the all vendor packages
	${VENDOR_CMD} update -all
	${MAKE} vendor-clean

vendor-clean:  ## Cleanup vendor packages "*_test" files, testdata and nogo files.
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

vendor-guru:  ## Update the internal guru package
	${RM} $(shell find ${PACKAGE_DIR}/src/nvim-go/internal/guru -maxdepth 1 -type f -name '*.go' -not -name 'result.go')
	${VENDOR_CMD} fetch golang.org/x/tools/cmd/guru
	mv ${PACKAGE_DIR}/vendor/src/golang.org/x/tools/cmd/guru/*.go ${PACKAGE_DIR}/src/nvim-go/internal/guru
	# Rename main to guru
	grep "package main" ${PACKAGE_DIR}/src/nvim-go/internal/guru/*.go -l | xargs sed -i 's/package main/package guru/'
	# Add Result interface
	sed -i "s|PrintPlain(printf printfFunc)|\0\n\n\tResult(fset *token.FileSet) interface{}|" ${PACKAGE_DIR}/src/nvim-go/internal/guru/guru.go
	# Export functions
	grep "findPackageMember" ${PACKAGE_DIR}/src/nvim-go/internal/guru/*.go -l | xargs sed -i 's/findPackageMember/FindPackageMember/'
	grep "packageForQualIdent" ${PACKAGE_DIR}/src/nvim-go/internal/guru/*.go -l | xargs sed -i 's/packageForQualIdent/PackageForQualIdent/'
	grep "guessImportPath" ${PACKAGE_DIR}/src/nvim-go/internal/guru/*.go -l | xargs sed -i 's/guessImportPath/GuessImportPath/'
	# ignore build main.go
	sed -i "s|package guru // import \"golang.org/x/tools/cmd/guru\"|\n// +build ignore\n\n\0|" ${PACKAGE_DIR}/src/nvim-go/internal/guru/main.go
	# ignore build guru_test.go
	sed -i "s|package guru_test|// +build ignore\n\n\0|" ${PACKAGE_DIR}/src/nvim-go/internal/guru/guru_test.go
	${VENDOR_CMD} delete golang.org/x/tools/cmd/guru
	${VENDOR_CMD} update golang.org/x/tools/cmd/guru/serial


docker: docker-run  ## Run the docker container test on Linux

docker-build:  ## Build the zchee/nvim-go docker container for testing on the Linux
	${DOCKER_CMD} build --rm -t ${GITHUB_USER}/${PACKAGE_NAME} .

docker-build-nocache:  ## Build the zchee/nvim-go docker container for testing on the Linux without cache
	${DOCKER_CMD} build --rm --no-cache -t ${GITHUB_USER}/${PACKAGE_NAME} .

docker-run: docker-build  ## Run the zchee/nvim-go docker container test
	${DOCKER_CMD} run --rm -it ${GITHUB_USER}/${PACKAGE_NAME} ${GO_TEST} $(GO_TEST_FLAGS)

clean:  ## Clean the {bin,pkg} directory
	${RM} -r ./bin ./pkg


todo:  ## Print the all of (TODO|BUG|XXX|FIXME|NOTE) in nvim-go package sources
	@pt -e 'TODO(\(.+\):|:)' --after=1 --ignore vendor --ignore internal --ignore Makefile || true
	@pt -e 'BUG(\(.+\):|:)' --after=1 --ignore vendor --ignore internal  --ignore Makefile || true
	@pt -e 'XXX(\(.+\):|:)' --after=1 --ignore vendor --ignore internal  --ignore Makefile || true
	@pt -e 'FIXME(\(.+\):|:)' --after=1 --ignore vendor --ignore internal --ignore Makefile || true
	@pt -e 'NOTE(\(.+\):|:)' --after=1 --ignore vendor --ignore internal --ignore Makefile || true

help:  ## Print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: clean test build build-race rebuild manifest test test-docker test-bench test-run vendor-all vendor-guru docker docker-build docker-build-nocache todo help

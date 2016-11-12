GITHUB_USER := zchee

ifeq ($(DEBUG),true)
	GO_GCFLAGS += -gcflags "-N -l"
else
	GO_LDFLAGS += -ldflags "-w -s"
endif

PACKAGE_NAME := nvim-go
PACKAGE_DIR := $(shell pwd)
BINARY_NAME := bin/nvim

CC := clang
CXX := clang++
GOPATH ?= $(shell go env GOPATH)
GOROOT ?= $(shell go env GOROOT)

GO_CMD := go
GB_CMD := gb
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
GO_BUILD_RACE := ${GB_CMD} build -race
GO_TEST := ${GB_CMD} test
GO_LINT := golint


default: build

build: $(PACKAGE_DIR)/plugin/manifest ## build the nvim-go binary
	${GO_BUILD} $(GO_LDFLAGS) ${GO_GCFLAGS} || exit 1
	$(PACKAGE_DIR)/plugin/manifest -w $(PACKAGE_NAME)

build-race: $(PACKAGE_DIR)/plugin/manifest ## build the nvim-go binary with -race flag
	${GO_BUILD} -race $(GO_LDFLAGS) ${GO_GCFLAGS} || exit 1
	mv ./bin/nvim-go-race ./bin/nvim-go
	$(PACKAGE_DIR)/plugin/manifest -w $(PACKAGE_NAME)

rebuild: clean $(PACKAGE_DIR)/plugin/manifest ## rebuild the nvim-go binary
	${GO_BUILD} -f $(GO_LDFLAGS) ${GO_GCFLAGS} || exit 1
	$(PACKAGE_DIR)/plugin/manifest -w $(PACKAGE_NAME)

$(PACKAGE_DIR)/plugin/manifest: ## build the auto writing neovim manifest utility binary
	$(GO_CMD) build -o $(PACKAGE_DIR)/plugin/manifest $(PACKAGE_DIR)/plugin/manifest.go

test: ## test the nvim-go package
	${GO_TEST} $(GO_TEST_FLAGS) || exit 1

test-docker: docker-run ## run the docker container test on Linux

test-bench: ## test the nvim-go package with benchmark
	${GO_TEST} $(GO_TEST_FLAGS) || exit 1

test-run: ## test the nvim-go package run the any function only
	${GO_TEST_RUN} || exit 1

vendor-all: ## update the all vendor packages
	${VENDOR_CMD} update -all

vendor-guru: ## update the internal guru package
	${RM} -r ${PACKAGE_DIR}/src/nvim-go/internal/guru/*.go
	${VENDOR_CMD} fetch golang.org/x/tools/cmd/guru
	cp -r ${PACKAGE_DIR}/vendor/src/golang.org/x/tools/cmd/guru ${PACKAGE_DIR}/src/nvim-go/internal/guru
	${RM} -r ${PACKAGE_DIR}/src/nvim-go/internal/guru/{main.go,*_test.go,serial,testdata,*.bash,*.vim,*.el}
	grep "package main" ${PACKAGE_DIR}/src/nvim-go/internal/guru/*.go -l | xargs sed -i 's/package main/package guru/'
	${VENDOR_CMD} delete golang.org/x/tools/cmd/guru

docker: docker-run ## run the docker container test on Linux

docker-build: ## build the zchee/nvim-go docker container for testing on the Linux
	${DOCKER_CMD} build --rm -t ${GITHUB_USER}/${PACKAGE_NAME} .

docker-build-nocache: ## build the zchee/nvim-go docker container for testing on the Linux without cache
	${DOCKER_CMD} build --rm --no-cache -t ${GITHUB_USER}/${PACKAGE_NAME} .

docker-run: docker-build ## run the zchee/nvim-go docker container test
	${DOCKER_CMD} run --rm -it ${GITHUB_USER}/${PACKAGE_NAME} ${GO_TEST} $(GO_TEST_FLAGS)

clean: ## clean the ./bin and ./pkg directory
	${RM} -r ./bin ./pkg

todo: ## print the all of TODO,BUG,XXX,FIXME,NOTE in nvim-go package sources
	@ag 'TODO(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal --ignore Makefile || true
	@ag 'BUG(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal  --ignore Makefile|| true
	@ag 'XXX(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal  --ignore Makefile|| true
	@ag 'FIXME(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal --ignore Makefile || true
	@ag 'NOTE(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal --ignore Makefile || true

help: ## print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: clean test build build-race rebuild test test-docker test-bench test-run vendor-all vendor-guru docker docker-build docker-build-nocache todo help

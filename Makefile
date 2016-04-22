GITHUB_USER := zchee

VERBOSE := -v
GIT_VERSION := ${GO_GCFLAGS} -X `go list ./version`.GitCommit=`git rev-parse --short HEAD 2>/dev/null`
ifeq ($(RELEASE),true)
	GO_LDFLAGS += -ldflags "-w -s"
else
	GO_GCFLAGS += -gcflags "-N -l"
endif

TOP_PACKAGE_DIR := github.com/${GITHUB_USER}
PACKAGE_NAME := $(shell basename $(PWD))
PACKAGE_DIR := ${HOME}/src/${TOP_PACKAGE_DIR}/${PACKAGE_NAME}
BINARY_NAME := bin/nvim-go-darwin-amd64

CC := clang
CXX := clang++
GO_CMD := gb
VENDOR_CMD := ${GO_CMD} vendor
DOCKER_CMD := docker

GO_LDFLAGS ?=
GO_GCFLAGS ?=
CGO_CFLAGS ?=
CGO_CPPFLAGS ?=
CGO_CXXFLAGS ?=
CGO_LDFLAGS ?=

GO_BUILD := ${GO_CMD} build
GO_BUILD_RACE := ${GO_CMD} build ${VERBOSE} -race
GO_TEST := ${GO_CMD} test ${VERBOSE}
GO_TEST_RUN := ${GO_TEST} -run ${RUN}
GO_TEST_ALL := test -race -cover -bench=.
GO_LINT := =golint
GO_VET=${GO_CMD} vet
GO_RUN=${GO_CMD} run
GO_INSTALL=${GO_CMD} install
GO_CLEAN=${GO_CMD} clean


default: build

build:
	${GO_BUILD} $(GO_LDFLAGS) ${GO_GCFLAGS} || exit 1

rebuild:
	${GO_BUILD} -f -F $(GO_LDFLAGS) ${GO_GCFLAGS} || exit 1

test:
	${GO_TEST} || exit 1

test/run:
	${GO_TEST_RUN} || exit 1

vendor/all:
	${VENDOR_CMD} update -all

vendor/guru:
	${RM} -r ${PACKAGE_DIR}/src/nvim-go/guru
	${VENDOR_CMD} fetch golang.org/x/tools/cmd/guru
	cp -r ${PACKAGE_DIR}/vendor/src/golang.org/x/tools/cmd/guru ${PACKAGE_DIR}/src/nvim-go/guru
	${RM} -r ${PACKAGE_DIR}/src/nvim-go/guru/{main.go,*_test.go,serial,testdata,*.bash,*.vim,*.el}
	grep "package main" ${PACKAGE_DIR}/src/nvim-go/guru/*.go -l | xargs sed -i 's/package main/package guru/'
	${VENDOR_CMD} delete golang.org/x/tools/cmd/guru
	${VENDOR_CMD} update golang.org/x/tools/cmd/guru/serial

docker/build:
	${DOCKER_CMD} build --rm -t ${GITHUB_USER}/${PACKAGE_NAME} .

docker/build-nocache:
	${DOCKER_CMD} build --rm --no-cache -t ${GITHUB_USER}/${PACKAGE_NAME} .

clean:
	@${RM} -r ./bin ./pkg

.PHONY: clean build test test-run dep-save dep-restore docker-build docker-build-nocache

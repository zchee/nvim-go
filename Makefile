GITHUB_USER := zchee

VERBOSE := -v
ifeq ($(DEBUG),true)
	GO_GCFLAGS += -gcflags "-N -l"
else
	GO_LDFLAGS += -ldflags "-w -s"
endif

TOP_PACKAGE_DIR := github.com/${GITHUB_USER}
PACKAGE_NAME := nvim-go
PACKAGE_DIR := ${GOPATH}/src/${TOP_PACKAGE_DIR}/${PACKAGE_NAME}
BINARY_NAME := bin/nvim

CC := clang
CXX := clang++
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

GO_BUILD := ${GB_CMD} build
GO_BUILD_RACE := ${GB_CMD} build -race
GO_TEST := ${GB_CMD} test ${VERBOSE}
GO_LINT := golint


default: build

build: $(PACKAGE_DIR)/plugin/specs
	${GO_BUILD} $(GO_LDFLAGS) ${GO_GCFLAGS} || exit 1
	$(PACKAGE_DIR)/plugin/specs -w $(PACKAGE_NAME)

rebuild: clean $(PACKAGE_DIR)/plugin/specs
	${GO_BUILD} -f $(GO_LDFLAGS) ${GO_GCFLAGS} || exit 1
	$(PACKAGE_DIR)/plugin/specs -w $(PACKAGE_NAME)

$(PACKAGE_DIR)/plugin/specs:
	$(GO_CMD) build -o $(PACKAGE_DIR)/plugin/specs $(PACKAGE_DIR)/plugin/specs.go

test:
	${GO_TEST} || exit 1

test/run:
	${GO_TEST_RUN} || exit 1

vendor/all:
	${VENDOR_CMD} update -all

vendor/guru:
	${RM} -r ${PACKAGE_DIR}/src/nvim-go/internal/guru
	${VENDOR_CMD} fetch golang.org/x/tools/cmd/guru
	cp -r ${PACKAGE_DIR}/vendor/src/golang.org/x/tools/cmd/guru ${PACKAGE_DIR}/src/nvim-go/internal/guru
	${RM} -r ${PACKAGE_DIR}/src/nvim-go/internal/guru/{main.go,*_test.go,serial,testdata,*.bash,*.vim,*.el}
	grep "package main" ${PACKAGE_DIR}/src/nvim-go/internal/guru/*.go -l | xargs sed -i 's/package main/package guru/'
	${VENDOR_CMD} delete golang.org/x/tools/cmd/guru
	${VENDOR_CMD} update golang.org/x/tools/cmd/guru/serial

docker/build:
	${DOCKER_CMD} build --rm -t ${GITHUB_USER}/${PACKAGE_NAME} .

docker/build-nocache:
	${DOCKER_CMD} build --rm --no-cache -t ${GITHUB_USER}/${PACKAGE_NAME} .

clean:
	${RM} -r ./bin ./pkg

todo: 
	@ag 'TODO(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal --ignore Makefile || true
	@ag 'BUG(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal  --ignore Makefile|| true
	@ag 'XXX(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal  --ignore Makefile|| true
	@ag 'FIXME(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal --ignore Makefile || true
	@ag 'NOTE(\(.+\):|:)' --after=1 --ignore-dir vendor --ignore-dir internal --ignore Makefile || true

.PHONY: clean test

.DEFAULT_GOAL := build

# ----------------------------------------------------------------------------
# package level setting

APP := $(notdir $(CURDIR))
PACKAGE_ROOT := $(CURDIR)
PACKAGES := $(shell go list ./pkg/...)
VENDOR_PACKAGES := $(shell go list -deps ./pkg/...)

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
GO_TEST_FUNCS ?= .
GO_TEST_FLAGS ?= -race -run=$(GO_TEST_FUNCS)
GO_BENCH_FUNCS ?= .
GO_BENCH_FLAGS ?= -bench=${GO_BENCH_FUNCS} -benchmem

ifneq ($(NVIM_GO_DEBUG),)
GO_GCFLAGS+=-gcflags=all="-N -l -dwarflocationlists=true"  # https://tip.golang.org/doc/diagnostics.html#debugging
else
GO_LDFLAGS+=-ldflags "-w -s"
endif

# ----------------------------------------------------------------------------
# targets

.PHONY: init
init:  ## Install dependency tools
	go get -u -v \
		github.com/golang/dep/cmd/dep \
		\
		github.com/kisielk/errcheck \
		github.com/mdempsky/unconvert \
		golang.org/x/lint/golint \
		honnef.co/go/tools/cmd/gosimple \
		honnef.co/go/tools/cmd/staticcheck \
		honnef.co/go/tools/cmd/unused \
		\
		github.com/rakyll/gotest

.PHONY: build
build:  ## Build the nvim-go binary
	go build -v -o ./bin/nvim-go $(strip ${GO_BUILD_FLAGS} ${GO_GCFLAGS} ${GO_LDFLAGS}) ./cmd/nvim-go

.PHONY: build.race
build.race: GO_BUILD_FLAGS+=-race
build.race: clean build ## Build the nvim-go binary with race

.PHONY: build.rebuild
build.rebuild: clean build  ## Rebuild the nvim-go binary

.PHONY: manifest
manifest: build  ## Write plugin manifest for developer
	./bin/${APP} -manifest ${APP} -location ./plugin/${APP}.vim > /dev/null 2>&1

.PHONY: manifest.race
manifest.race: APP=${APP}-race
manifest.race: build.race  ## Write plugin manifest for developer
	./bin/${APP} -manifest ${APP} -location ./plugin/${APP}.vim > /dev/null 2>&1

.PHONY: manifest.dump
manifest.dump: build  ## Dump plugin manifest
	./bin/${APP} -manifest ${APP} 2>/dev/null


.PHONY: test
test:  ## Run the package test
	${GO_TEST} -v $(strip ${GO_TEST_FLAGS} ${PACKAGES})

.PHONY: bench
bench: GO_TEST_FUNCS=^$$
bench: GO_TEST_FLAGS+=${GO_BENCH_FLAGS}
bench: test ## Take the packages benchmark


.PHONY: lint
lint: lint.golint lint.errcheck lint.gosimple lint.staticcheck lint.unparam lint.vet  ## Run lint use all tools

.PHONY: lint.golint
lint.golint:  ## Run golint
	@echo "+ $@"
	@golint -set_exit_status -min_confidence=0.6 ${PACKAGES}

.PHONY: lint.errcheck
lint.errcheck:  ## Run errcheck
	@echo "+ $@"
	@errcheck ${PACKAGES}

.PHONY: lint.gosimple
lint.gosimple:  ## Run gosimple
	@echo "+ $@"
	@gosimple ${PACKAGES}

.PHONY: lint.staticcheck
lint.staticcheck:  ## Run staticcheck
	@echo "+ $@"
	@staticcheck $(PACKAGES)

.PHONY: lint.unparam
lint.unused:  ## Run unused
	@echo "+ $@"
	@unused $(PACKAGES)

.PHONY: lint.vet
lint.vet:  ## Run go vet
	@echo "+ $@"
	@go vet -all -shadow ${PACKAGES}

.PHONY: lint.unconvert
lint.unconvert:  ## Run unconvert
	@echo "+ $@"
	@unconvert -v ${PACKAGES}

.PHONY: coverage
coverage:  # take test coverage
	${GO_TEST} -v -race -covermode=atomic -coverprofile=$@.out -coverpkg=./pkg/... $(PACKAGES)


.PHONY: vendor.update
vendor.update:  ## Update the all vendor packages
	dep ensure -v -update

.PHONY: vendor.install
vendor.install:  # Install vendor packages for gocode completion
	go install -v -x ${VENDOR_PACKAGES}

.PHONY: vendor.guru
vendor.guru: vendor.guru-update vendor.guru-rename

.PHONY: vendor.guru-update
vendor.guru-update:  ## Update the internal guru package
	sed -i 's|unused-packages|# unused-packages|' Gopkg.toml
	dep ensure -v -update golang.org/x/tools
	${RM} -r $(shell find ${PACKAGE_ROOT}/pkg/internal/guru -maxdepth 1 -type f -name '*.go' -not -name 'result.go')
	cp ${PACKAGE_ROOT}/vendor/golang.org/x/tools/cmd/guru/*.go ${PACKAGE_ROOT}/pkg/internal/guru
	sed -i "s|\t// TODO(adonovan): opt: parallelize.|\tbp.GoFiles = append(bp.GoFiles, bp.CgoFiles...)\n\n\0|" pkg/internal/guru/definition.go
	sed -i 's| // import "golang.org/x/tools/cmd/guru"||' ./pkg/internal/guru/main.go
	sed -i 's|# unused-packages|unused-packages|' Gopkg.toml
	export DEP_REVISION=$(dep status -detail -f='{{range $$i, $$p := .Projects}}{{if eq $$p.ProjectRoot "golang.org/x/tools"}}{{$$p.Locked.Revision}}{{end}}{{end}}')
	perl -i -0pe 's|  name = "golang.org/x/tools"\n  branch = "master"\n|  name = "golang.org/x/tools"\n  revision = "${DEP_REVISION}"\n|m' Gopkg.toml
	dep ensure -v -vendor-only
	perl -i -0pe 's|  name = "golang.org/x/tools"\n  revision = "${DEP_REVISION}"\n|  name = "golang.org/x/tools"\n  branch = "master"\n|m' Gopkg.toml
	dep ensure -v -no-vendor
	unset DEP_REVISION

.PHONY: vendor.guru-rename
vendor.guru-rename: vendor.guru-update
	@echo -e "[INFO] Rename main to guru\\n"
	grep "package main" ${PACKAGE_ROOT}/pkg/internal/guru/*.go -l | xargs sed -i 's/package main/package guru/'
	@echo -e "[INFO] Add Result interface\\n"
	sed -i "s|PrintPlain(printf printfFunc)|\0\n\n\tResult(fset *token.FileSet) interface{}|" ${PACKAGE_ROOT}/pkg/internal/guru/guru.go
	@echo -e "[INFO] Export functions\\n"
	grep "findPackageMember" ${PACKAGE_ROOT}/pkg/internal/guru/*.go -l | xargs sed -i 's/findPackageMember/FindPackageMember/'
	grep "packageForQualIdent" ${PACKAGE_ROOT}/pkg/internal/guru/*.go -l | xargs sed -i 's/packageForQualIdent/PackageForQualIdent/'
	grep "guessImportPath" ${PACKAGE_ROOT}/pkg/internal/guru/*.go -l | xargs sed -i 's/guessImportPath/GuessImportPath/'
	@echo -e "[INFO] remove canonical custom import path from main.go\\n"
	sed -i "s|package guru|\n// +build ignore\n\n\0|" ${PACKAGE_ROOT}/pkg/internal/guru/main.go


.PHONY: clean
clean:  ## Clean the {bin,pkg} directory
	${RM} -r ./bin *.out *.prof


.PHONY: docker
docker: docker.test  ## Run the docker container test on Linux

.PHONY: docker.build
docker.build:  ## Build the zchee/nvim-go docker container for testing on the Linux
	docker build --rm -t ${USER}/${APP} .

.PHONY: docker.build-nocache
docker.build-nocache:  ## Build the zchee/nvim-go docker container for testing on the Linux without cache
	docker build --rm --no-cache -t ${USER}/${APP} .

.PHONY: docker.test
docker.test: docker-build  ## Run the package test with docker container
	docker run --rm -it ${USER}/${APP} go test -v ${GO_TEST_FLAGS} ${PACKAGES}


.PHONY: todo
todo:  ## Print the all of (TODO|BUG|XXX|FIXME|NOTE) in nvim-go package sources
	@pt -e '(TODO|BUG|XXX|FIXME|NOTE)(\(.+\):|:)' --follow --hidden --ignore=.git --ignore=vendor --ignore=internal --ignore=Makefile --ignore=snippets --ignore=indent

.PHONY: help
help:  ## Print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z./_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

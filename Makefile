.DEFAULT_GOAL := build

# ----------------------------------------------------------------------------
# package level setting

APP := $(notdir $(CURDIR))
PACKAGE_ROOT := $(CURDIR)
PACKAGES := $(shell go list ./pkg/... | grep -v -e 'pkg/internal/go' -e 'pkg/internal/gotool')
VENDOR_PACKAGES := $(shell go list -deps ./pkg/...)

GIT_TAG := $(shell git describe --tags --abbrev=0)
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# ----------------------------------------------------------------------------
# common environment variables

SHELL := /usr/bin/env bash
CC := clang  # need compile cgo for delve
CXX := clang++
GO_LDFLAGS ?= -X=main.tag=$(GIT_TAG) -X=main.gitCommit=$(GIT_COMMIT)
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
GO_GCFLAGS+=all="-N -l -dwarflocationlists=true"  # https://tip.golang.org/doc/diagnostics.html#debugging
else
GO_LDFLAGS+=-w -s
endif

# ----------------------------------------------------------------------------
# targets

define target
	@printf "+ \\033[32m$(shell printf $@ | cut -d '.' -f2)\\033[0m\\n"
endef

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
	$(call target)
	go build -v -o ./bin/${APP} $(strip ${GO_BUILD_FLAGS}) -gcflags=$(strip ${GO_GCFLAGS}) -ldflags="$(strip ${GO_LDFLAGS})" ./cmd/nvim-go

.PHONY: build.race
build.race: GO_BUILD_FLAGS+=-race
build.race: clean build  ## Build the nvim-go binary with race
	$(call target)

.PHONY: build.rebuild
build.rebuild: clean build  ## Rebuild the nvim-go binary
	$(call target)

.PHONY: manifest
manifest: build  ## Write plugin manifest for developer
	$(call target)
	./bin/${APP} -manifest ${APP} -location ./plugin/nvim-go.vim

.PHONY: manifest.race
manifest.race: APP=nvim-race
manifest.race: build.race manifest  ## Write plugin manifest for developer
	$(call target)

.PHONY: manifest.dump
manifest.dump: build  ## Dump plugin manifest
	$(call target)
	./bin/${APP} -manifest ${APP}


.PHONY: vendor.push
vendor.push:
	$(call target)
	sed -i 's|# unused-packages|unused-packages|' Gopkg.toml
	dep ensure -v
	git add Gopkg* vendor
	sed -i 's|unused-packages|# unused-packages|' Gopkg.toml
	dep ensure -v

.PHONY: vendor.update
vendor.update:  ## Update the all vendor packages
	$(call target)
	dep ensure -v -update

.PHONY: vendor.install
vendor.install:  # Install vendor packages for gocode completion
	$(call target)
	go install -v -x ${VENDOR_PACKAGES}


.PHONY: vendor.guru
vendor.guru: vendor.guru-update vendor.guru-rename
	$(call target)

.PHONY: vendor.guru-update
vendor.guru-update:  ## Update the internal guru package
	$(call target)
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
	$(call target)
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


.PHONY: test
test:  ## Run the package test
	$(call target)
	${GO_TEST} -v $(strip ${GO_TEST_FLAGS} ${PACKAGES})

.PHONY: bench
bench: GO_TEST_FUNCS=^$$
bench: GO_TEST_FLAGS+=${GO_BENCH_FLAGS}
bench: test ## Take the packages benchmark
	$(call target)


.PHONY: lint
lint: lint.golint lint.errcheck lint.megacheck lint.vet lint.unconvert  ## Run lint use all tools

.PHONY: lint.golint
lint.golint:  ## Run golint
	$(call target)
	golint -set_exit_status -min_confidence=0.6 ${PACKAGES}

.PHONY: lint.errcheck
lint.errcheck:  ## Run errcheck
	$(call target)
	errcheck ${PACKAGES}

.PHONY: lint.megacheck
lint.megacheck:  ## Run megacheck
	$(call target)
	megacheck $(PACKAGES)

.PHONY: lint.vet
lint.vet:  ## Run go vet
	$(call target)
	go vet -all ${PACKAGES}

.PHONY: lint.unconvert
lint.unconvert:  ## Run unconvert
	$(call target)
	unconvert -v ${PACKAGES}

.PHONY: coverage
coverage:  # take test coverage
	$(call target)
	${GO_TEST} -v -race -covermode=atomic -coverprofile=$@.out -coverpkg=./pkg/... $(PACKAGES)


.PHONY: clean
clean:  ## Clean the {bin,pkg} directory
	$(call target)
	${RM} -r ./bin *.out *.prof


.PHONY: docker
docker: docker.test  ## Run the docker container test on Linux
	$(call target)

.PHONY: docker.build
docker.build:  ## Build the zchee/nvim-go docker container for testing on the Linux
	$(call target)
	docker build --rm -t ${USER}/${APP} .

.PHONY: docker.build-nocache
docker.build-nocache:  ## Build the zchee/nvim-go docker container for testing on the Linux without cache
	$(call target)
	docker build --rm --no-cache -t ${USER}/${APP} .

.PHONY: docker.test
docker.test: docker-build  ## Run the package test with docker container
	$(call target)
	docker run --rm -it ${USER}/${APP} go test -v ${GO_TEST_FLAGS} ${PACKAGES}


.PHONY: todo
todo:  ## Print the all of (TODO|BUG|XXX|FIXME|NOTE) in nvim-go package sources
	@pt -e '(TODO|BUG|XXX|FIXME|NOTE)(\(.+\):|:)' --follow --hidden --ignore=.git --ignore=vendor --ignore=internal --ignore=Makefile --ignore=snippets --ignore=indent

.PHONY: help
help:  ## Print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z./_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

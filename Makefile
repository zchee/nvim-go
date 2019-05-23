# ----------------------------------------------------------------------------
# global

.DEFAULT_GOAL = manifest
APP = nvim-go
CMD = $(PKG)/cmd/$(APP)
CGO_ENABLED = 1

# ----------------------------------------------------------------------------
# target

# manifest

.PHONY: manifest
manifest: build  ## Writes the plugin manifest.
	$(call target)
	@$(CURDIR)/bin/${APP} -manifest ${APP} -location $(CURDIR)/plugin/nvim-go.vim

.PHONY: manifest/race
manifest/race: build/race manifest  ## Writes plugin manifest with the race prefix.
	$(call target)

.PHONY: manifest/dump
manifest/dump: build  ## Dumps plugin manifest.
	$(call target)
	@$(CURDIR)/bin/${APP} -manifest ${APP}

# mod

.PHONY: mod/lock/delve
mod/lock/delve:  ## Locks version with go-delve/delve@92dad94.
	$(call target)
	@go get -u -m -v -x github.com/derekparker/delve@92dad94
	@go get -u -m -v -x golang.org/x/arch@f4009597
	@go get -u -m -v -x golang.org/x/debug@fb50892
	@go get -u -m -v -x golang.org/x/sys@f3918c30c

.PHONY: mod/install
mod/install: mod/tidy mod/vendor

.PHONY: mod/update
mod/update: mod/get mod/lock/delve mod/tidy mod/vendor mod/install

.PHONY: mod
mod: mod/init mod/lock/delve mod/tidy mod/vendor mod/install

# internal vendor packages

## guru
.PHONY: vendor/guru/update
vendor/guru/update:
	$(call target)
	@GO111MODULE=on go get -u -m -v golang.org/x/tools@master
	printf "%s\\n\\n%s" 'package guru' 'import _ "golang.org/x/tools/cmd/guru"' > $(CURDIR)/pkg/internal/guru/hack.go
	@GO111MODULE=on go mod vendor -v
	${RM} -r $(shell find $(CURDIR)/pkg/internal/guru -maxdepth 1 -type f -name '*.go' -not -name 'result.go')
	mv $(CURDIR)/vendor/golang.org/x/tools/cmd/guru/*.go $(CURDIR)/pkg/internal/guru
	sed -i "s|\t// TODO(adonovan): opt: parallelize.|\tbp.GoFiles = append(bp.GoFiles, bp.CgoFiles...)\n\n\0|" $(CURDIR)/pkg/internal/guru/definition.go
	sed -i 's| // import "golang.org/x/tools/cmd/guru"||' $(CURDIR)/pkg/internal/guru/main.go

.PHONY: vendor/guru/rename
vendor/guru/rename: vendor/guru/update
	$(call target)
	grep "package main" $(CURDIR)/pkg/internal/guru/*.go -l | xargs sed -i 's/package main/package guru/'
	sed -i "s|PrintPlain(printf printfFunc)|\0\n\n\tResult(fset *token.FileSet) interface{}|" $(CURDIR)/pkg/internal/guru/guru.go
	grep "findPackageMember" $(CURDIR)/pkg/internal/guru/*.go -l | xargs sed -i 's/findPackageMember/FindPackageMember/'
	grep "packageForQualIdent" $(CURDIR)/pkg/internal/guru/*.go -l | xargs sed -i 's/packageForQualIdent/PackageForQualIdent/'
	grep "guessImportPath" $(CURDIR)/pkg/internal/guru/*.go -l | xargs sed -i 's/guessImportPath/GuessImportPath/'
	sed -i "s|package guru|\n// +build ignore\n\n\0|" $(CURDIR)/pkg/internal/guru/main.go

.PHONY: vendor/guru
vendor/guru: vendor/guru/update vendor/guru/rename mod/install
vendor/guru:  ## Updates the vendoring guru package into pkg/internal.

## golang.org/x/tools/internal
.PHONY: vendor/x/internal/tools/update
vendor/x/tools/internal/update:
	@GO111MODULE=off go get -u -v golang.org/x/tools/internal/...

.PHONY: vendor/x/tools/internal/%
vendor/x/tools/internal/%:
	mkdir -p $(CURDIR)/pkg/internal/$*
	find $(CURDIR)/pkg/internal/$* -type f -name '*.go' -print -delete
	find /Users/zchee/go/src/golang.org/x/tools/internal/$* -type f -name '*.go' -and -not -name '*_test.go' -exec cp {} $(CURDIR)/pkg/internal/$* \;

.PHONY: vendor/x/tools/internal
vendor/x/tools/internal: vendor/x/tools/internal/update
vendor/x/tools/internal:  ## Updates the vendoring golang.org/x/tools/internal packages into pkg/internal.
	${MAKE} vendor/x/tools/internal/fastwalk vendor/x/tools/internal/gopathwalk
	sed -i "s|golang.org/x/tools/internal/fastwalk|$(PKG)/pkg/internal/fastwalk|" $(CURDIR)/pkg/internal/gopathwalk/walk.go

## github.com/valyala/bytebufferpool
.PHONY: vendor/bytebufferpool/update
vendor/bytebufferpool/update:
	@GO111MODULE=off go get -u -v github.com/valyala/bytebufferpool

.PHONY: vendor/x/tools
vendor/bytebufferpool: vendor/bytebufferpool/update
vendor/bytebufferpool:  ## Update vendoring valyala/bytebufferpool package into pkg/internal.
	mkdir -p $(CURDIR)/pkg/internal/$(subst vendor/bytebuffer,,$@)
	find $(CURDIR)/pkg/internal/$(subst vendor/bytebuffer,,$@) -type f -name '*.go' -print -delete
	find $(GO_PATH)/src/github.com/$(subst vendor,valyala,$@) -type f -name '*.go' -and -not -name '*_test.go' -exec cp {} $(CURDIR)/pkg/internal/$(subst vendor/bytebuffer,,$@) \;


## miscellaneous

.PHONY: container/test
container/test: container/build  ## Runs package test into Linux container.
	$(call target)
	docker container run ${CONTAINER_RUN_ARGS} $(IMAGE_REGISTRY)/$(APP):$(VERSION:v%=%) $(GO_TEST) -v -race $(strip $(GO_FLAGS)) -run=$(GO_TEST_FUNC) $(GO_TEST_PKGS)

.PHONY: container/bench
container/bench: container/build  ## Runs package benchmarks into Linux container.
	$(call target)
	docker container run ${CONTAINER_RUN_ARGS} $(IMAGE_REGISTRY)/$(APP):$(VERSION:v%=%) $(GO_TEST) -v $(strip $(GO_FLAGS)) -run='^$$' -bench=$(GO_BENCH_FUNC) -benchmem $(GO_TEST_PKGS)

# ----------------------------------------------------------------------------
# include

include hack/make/go.mk

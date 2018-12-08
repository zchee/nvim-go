# ----------------------------------------------------------------------------
# global

APP = nvim-go

# ---------------------------------------------------------------------------
# target

# manifest
.PHONY: manifest
manifest: build  ## Write plugin manifest for developer
	$(call target)
	@$(CURDIR)/bin/${APP} -manifest ${APP} -location $(CURDIR)/plugin/nvim-go.vim

.PHONY: manifest/race
manifest/race: APP=nvim-race
manifest/race: build/race manifest  ## Write plugin manifest for developer
	$(call target)

.PHONY: manifest/dump
manifest/dump: build  ## Dump plugin manifest
	$(call target)
	@$(CURDIR)/bin/${APP} -manifest ${APP}


# internal vendor
.PHONY: vendor/guru/update
vendor/guru/update:  ## Update the internal guru package
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

.PHONY: vendor/guru/rename
vendor/guru/rename: vendor/guru/update
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

.PHONY: vendor/guru
vendor/guru: vendor/guru/update vendor/guru/rename
	$(call target)

.PHONY: vendor/x/tools/update
vendor/x/tools/internal/update:
	@go get -u -v golang.org/x/tools/internal/...

.PHONY: vendor/x/tools/%
vendor/x/tools/%:
	mkdir -p ${PACKAGE_ROOT}/pkg/internal/$*
	find ${PACKAGE_ROOT}/pkg/internal/$* -type f -name '*.go' -print -delete
	find /Users/zchee/go/src/golang.org/x/tools/internal/$* -type f -name '*.go' -and -not -name '*_test.go' -exec cp {} ${PACKAGE_ROOT}/pkg/internal/$* \;

.PHONY: vendor/x/tools
vendor/x/tools: vendor/x/tools/update vendor/x/tools/fastwalk vendor/x/tools/gopathwalk vendor/x/tools/semver
	sed -i "s|golang.org/x/tools/internal/fastwalk|github.com/zchee/nvim-go/pkg/internal/fastwalk|" ${PACKAGE_ROOT}/pkg/internal/gopathwalk/walk.go

.PHONY: vendor/bytebufferpool/update
vendor/bytebufferpool/update:
	@go get -u -v github.com/valyala/bytebufferpool

.PHONY: vendor/x/tools
vendor/bytebufferpool: vendor/bytebufferpool/update
	mkdir -p ${PACKAGE_ROOT}/pkg/internal/$(subst vendor/bytebuffer,,$@)
	find ${PACKAGE_ROOT}/pkg/internal/$(subst vendor/bytebuffer,,$@) -type f -name '*.go' -print -delete
	find /Users/zchee/go/src/github.com/valyala/bytebufferpool -type f -name '*.go' -and -not -name '*_test.go' -exec cp {} ${PACKAGE_ROOT}/pkg/internal/$(subst vendor/bytebuffer,,$@) \;


.PHONY: docker
docker: docker/test  ## Run the docker container test on Linux
	$(call target)

.PHONY: docker/build
docker/build:  ## Build the zchee/nvim-go docker container for testing on the Linux
	$(call target)
	docker image build --rm --progress=plain -t $(IMAGE_REGISTRY)/$(APP) .

.PHONY: docker/build-nocache
docker/build-nocache:  ## Build the zchee/nvim-go docker container for testing on the Linux without cache
	$(call target)
	docker image build --rm --no-cache --progress=plain -t $(IMAGE_REGISTRY)/$(APP) .

.PHONY: docker/test
docker/test: docker/build  ## Run the package test with docker container
	$(call target)
	docker container run --rm -it $(IMAGE_REGISTRY)/$(APP) go test -v ${GO_TEST_FLAGS} ${PACKAGES}

# ----------------------------------------------------------------------------
# include

include hack/make/go.mk

# ----------------------------------------------------------------------------
# override

build: GOFLAGS+=-mod=vendor

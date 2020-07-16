export GO111MODULE=on

# By default, the target OS is the same as the host OS,
# but this can be overridden by setting TARGET_OS to "windows"/"darwin"/"linux".
GOOS:=$(shell go env GOOS)
TARGET_OS?=${GOOS}

## Build
.PHONY: build build-nogui check-has-go

BRIDGE_VERSION?=$(shell git describe --abbrev=0 --tags)-git
REVISION:=$(shell git rev-parse --short=10 HEAD)
BUILD_TIME:=$(shell date +%FT%T%z)

BUILD_TAGS?=pmapi_prod
BUILD_FLAGS:=-tags='${BUILD_TAGS}'
BUILD_FLAGS_NOGUI:=-tags='${BUILD_TAGS} nogui'
GO_LDFLAGS:=$(addprefix -X github.com/ProtonMail/proton-bridge/pkg/constants.,Version=${BRIDGE_VERSION} Revision=${REVISION} BuildTime=${BUILD_TIME})
ifneq "${BUILD_LDFLAGS}" ""
    GO_LDFLAGS+= ${BUILD_LDFLAGS}
endif
GO_LDFLAGS:=-ldflags '${GO_LDFLAGS}'
BUILD_FLAGS+= ${GO_LDFLAGS}
BUILD_FLAGS_NOGUI+= ${GO_LDFLAGS}

DEPLOY_DIR:=cmd/Desktop-Bridge/deploy
ICO_FILES:=
EXE:=$(shell basename ${CURDIR})

ifeq "${TARGET_OS}" "windows"
    EXE:=${EXE}.exe
    ICO_FILES:=logo.ico icon.rc icon_windows.syso
endif
ifeq "${TARGET_OS}" "darwin"
    DARWINAPP_CONTENTS:=${DEPLOY_DIR}/darwin/${EXE}.app/Contents
    EXE:=${EXE}.app/Contents/MacOS/${EXE}
endif
EXE_TARGET:=${DEPLOY_DIR}/${TARGET_OS}/${EXE}
TGZ_TARGET:=bridge_${TARGET_OS}_${REVISION}.tgz


build: ${TGZ_TARGET}

build-nogui:
	go build ${BUILD_FLAGS_NOGUI} -o Desktop-Bridge cmd/Desktop-Bridge/main.go

${TGZ_TARGET}: ${DEPLOY_DIR}/${TARGET_OS}
	rm -f $@
	cd ${DEPLOY_DIR} && tar czf ../../../$@ ${TARGET_OS}

${DEPLOY_DIR}/linux: ${EXE_TARGET}
	cp -pf ./internal/frontend/share/icons/logo.svg ${DEPLOY_DIR}/linux/
	cp -pf ./LICENSE ${DEPLOY_DIR}/linux/
	cp -pf ./Changelog.md ${DEPLOY_DIR}/linux/

${DEPLOY_DIR}/darwin: ${EXE_TARGET}
	cp ./internal/frontend/share/icons/Bridge.icns ${DARWINAPP_CONTENTS}/Resources/
	cp -r "utils/addcert.scpt" ${DARWINAPP_CONTENTS}/Resources/
	cp LICENSE ${DARWINAPP_CONTENTS}/Resources/
	rm -rf "${DARWINAPP_CONTENTS}/Frameworks/QtWebEngine.framework"
	rm -rf "${DARWINAPP_CONTENTS}/Frameworks/QtWebView.framework"
	rm -rf "${DARWINAPP_CONTENTS}/Frameworks/QtWebEngineCore.framework"
	./utils/remove_non_relative_links_darwin.sh "${EXE_TARGET}"

${DEPLOY_DIR}/windows: ${EXE_TARGET}
	cp ./internal/frontend/share/icons/logo.ico ${DEPLOY_DIR}/windows/
	cp LICENSE ${DEPLOY_DIR}/windows/

QT_BUILD_TARGET:=build desktop
ifneq "${GOOS}" "${TARGET_OS}"
  ifeq "${TARGET_OS}" "windows"
    QT_BUILD_TARGET:=-docker build windows_64_shared
  endif
endif

${EXE_TARGET}: check-has-go gofiles ${ICO_FILES} update-vendor
	rm -rf deploy ${TARGET_OS} ${DEPLOY_DIR}
	cp cmd/Desktop-Bridge/main.go .
	qtdeploy ${BUILD_FLAGS} ${QT_BUILD_TARGET}
	mv deploy cmd/Desktop-Bridge
	rm -rf ${TARGET_OS} main.go

logo.ico: ./internal/frontend/share/icons/logo.ico
	cp $^ .
icon.rc: ./internal/frontend/share/icon.rc
	cp $^ .
./internal/frontend/qt/icon_windows.syso: ./internal/frontend/share/icon.rc  logo.ico 
	windres --target=pe-x86-64 -o $@ $<
icon_windows.syso: ./internal/frontend/qt/icon_windows.syso
	cp $^ .


## Rules for therecipe/qt
.PHONY: prepare-vendor update-vendor
THERECIPE_ENV:=github.com/therecipe/env_${TARGET_OS}_amd64_513

# vendor folder will be deleted by gomod hence we cache the big repo
# therecipe/env in order to download it only once
vendor-cache/${THERECIPE_ENV}:
	git clone https://${THERECIPE_ENV}.git vendor-cache/${THERECIPE_ENV}

# The command used to make symlinks is different on windows.
# So if the GOOS is windows and we aren't crossbuilding (in which case the host os would still be *nix)
# we need to change the LINKCMD to something windowsy.
LINKCMD:=ln -sf ${CURDIR}/vendor-cache/${THERECIPE_ENV} vendor/${THERECIPE_ENV}
ifeq "${GOOS}" "windows"
  WINDIR:=$(subst /c/,c:\\,${CURDIR})/vendor-cache/${THERECIPE_ENV}
  LINKCMD:=cmd //c 'mklink $(subst /,\,vendor\${THERECIPE_ENV} ${WINDIR})'
endif

prepare-vendor:
	go install -v -tags=no_env github.com/therecipe/qt/cmd/...
	go mod vendor

# update-vendor is PHONY because we need to make sure that we always have updated vendor
update-vendor: vendor-cache/${THERECIPE_ENV} prepare-vendor
	${LINKCMD}


## Dev dependencies
.PHONY: install-devel-tools install-linter install-go-mod-outdated
LINTVER:="v1.27.0"
LINTSRC:="https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"

install-dev-dependencies: install-devel-tools install-linter install-go-mod-outdated

install-devel-tools: check-has-go
	go get -v github.com/golang/mock/gomock
	go get -v github.com/golang/mock/mockgen
	go get -v github.com/go-delve/delve

install-linter: check-has-go
	curl -sfL $(LINTSRC) | sh -s -- -b $(shell go env GOPATH)/bin $(LINTVER)

install-go-mod-outdated:
	which go-mod-outdated || go get -u github.com/psampaz/go-mod-outdated


## Checks, mocks and docs
.PHONY: check-has-go add-license change-copyright-year test bench coverage mocks lint-license lint-golang lint updates doc
check-has-go:
	@which go || (echo "Install Go-lang!" && exit 1)

add-license:
	./utils/missing_license.sh add

change-copyright-year:
	./utils/missing_license.sh change-year

test: gofiles
	@# Listing packages manually to not run Qt folder (which needs to run qtsetup first) and integration tests.
	go test -coverprofile=/tmp/coverage.out -run=${TESTRUN} \
		./internal/api/... \
		./internal/bridge/... \
		./internal/events/... \
		./internal/frontend/autoconfig/... \
		./internal/frontend/cli/... \
		./internal/imap/... \
		./internal/metrics/... \
		./internal/preferences/... \
		./internal/smtp/... \
		./internal/store/... \
		./internal/users/... \
		./pkg/...

bench:
	go test -run '^$$' -bench=. -memprofile bench_mem.pprof -cpuprofile bench_cpu.pprof ./internal/store
	go tool pprof -png -output bench_mem.png bench_mem.pprof
	go tool pprof -png -output bench_cpu.png bench_cpu.pprof

coverage: test
	go tool cover -html=/tmp/coverage.out -o=coverage.html

mocks:
	mockgen --package mocks github.com/ProtonMail/proton-bridge/internal/users Configer,PanicHandler,ClientManager,CredentialsStorer,StoreMaker > internal/users/mocks/mocks.go
	mockgen --package mocks github.com/ProtonMail/proton-bridge/internal/store PanicHandler,ClientManager,BridgeUser > internal/store/mocks/mocks.go
	mockgen --package mocks github.com/ProtonMail/proton-bridge/pkg/listener Listener > internal/store/mocks/utils_mocks.go
	mockgen --package mocks github.com/ProtonMail/proton-bridge/pkg/pmapi Client > pkg/pmapi/mocks/mocks.go

lint: lint-golang lint-license

lint-license:
	./utils/missing_license.sh check

lint-golang:
	which golangci-lint || $(MAKE) install-linter
	golangci-lint run ./...

updates: install-go-mod-outdated
	# Uncomment the "-ci" to fail the job if something can be updated.
	go list -u -m -json all | go-mod-outdated -update -direct #-ci

doc:
	godoc -http=:6060

.PHONY: gofiles
# Following files are for the whole app so it makes sense to have them in bridge package.
# (Options like cmd or internal were considered and bridge package is the best place for them.)
gofiles: ./internal/bridge/credits.go ./internal/bridge/release_notes.go
./internal/bridge/credits.go: ./utils/credits.sh go.mod
	cd ./utils/ && ./credits.sh
./internal/bridge/release_notes.go: ./utils/release-notes.sh ./release-notes/notes.txt ./release-notes/bugs.txt
	cd ./utils/ && ./release-notes.sh


## Run and debug
.PHONY: run run-qt run-qt-cli run-nogui run-nogui-cli run-debug qmlpreview qt-fronted-clean clean
VERBOSITY?=debug-client
RUN_FLAGS:=-m -l=${VERBOSITY}

run: run-nogui-cli

run-qt: ${EXE_TARGET}
	PROTONMAIL_ENV=dev ./$< ${RUN_FLAGS} | tee last.log
run-qt-cli: ${EXE_TARGET}
	PROTONMAIL_ENV=dev ./$< ${RUN_FLAGS} -c

run-nogui: clean-vendor gofiles
	PROTONMAIL_ENV=dev go run ${BUILD_FLAGS_NOGUI} cmd/Desktop-Bridge/main.go ${RUN_FLAGS} | tee last.log
run-nogui-cli: clean-vendor gofiles
	PROTONMAIL_ENV=dev go run ${BUILD_FLAGS_NOGUI} cmd/Desktop-Bridge/main.go ${RUN_FLAGS} -c

run-debug:
	PROTONMAIL_ENV=dev dlv debug --build-flags "${BUILD_FLAGS_NOGUI}" cmd/Desktop-Bridge/main.go -- ${RUN_FLAGS}

run-qml-preview:
	make -C internal/frontend/qt -f Makefile.local qmlpreview

clean-frontend-qt:
	make -C internal/frontend/qt -f Makefile.local clean

clean-vendor: clean-frontend-qt
	rm -rf ./vendor

clean: clean-frontend-qt
	rm -rf vendor-cache
	rm -rf cmd/Desktop-Bridge/deploy
	rm -f build last.log mem.pprof
	rm -rf logo.ico icon.rc icon_windows.syso internal/frontend/qt/icon_windows.syso

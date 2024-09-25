export GO111MODULE=on
export CGO_ENABLED=1

# By default, the target OS is the same as the host OS,
# but this can be overridden by setting TARGET_OS to "windows"/"darwin"/"linux".
GOOS:=$(shell go env GOOS)
TARGET_CMD?=Desktop-Bridge
TARGET_OS?=${GOOS}
ROOT_DIR:=$(realpath .)

## Build
.PHONY: build build-gui build-nogui build-launcher versioner hasher

# Keep version hardcoded so app build works also without Git repository.
BRIDGE_APP_VERSION?=3.14.0+git
APP_VERSION:=${BRIDGE_APP_VERSION}
APP_FULL_NAME:=Proton Mail Bridge
APP_VENDOR:=Proton AG
SRC_ICO:=bridge.ico
SRC_ICNS:=Bridge.icns
SRC_SVG:=bridge.svg
EXE_NAME:=proton-bridge
REVISION:=$(shell "${ROOT_DIR}/utils/get_revision.sh" rev)
TAG:=$(shell "${ROOT_DIR}/utils/get_revision.sh" tag)
BUILD_TIME:=$(shell date +%FT%T%z)
MACOS_MIN_VERSION_ARM64=11.0
MACOS_MIN_VERSION_AMD64=10.15
BUILD_ENV?=dev

BUILD_FLAGS:=-tags='${BUILD_TAGS}'
BUILD_FLAGS_LAUNCHER:=${BUILD_FLAGS}
GO_LDFLAGS:=$(addprefix -X github.com/ProtonMail/proton-bridge/v3/internal/constants., Version=${APP_VERSION} Revision=${REVISION} Tag=${TAG} BuildTime=${BUILD_TIME})
GO_LDFLAGS+=-X "github.com/ProtonMail/proton-bridge/v3/internal/constants.FullAppName=${APP_FULL_NAME}"

ifneq "${DSN_SENTRY}" ""
	GO_LDFLAGS+=-X github.com/ProtonMail/proton-bridge/v3/internal/constants.DSNSentry=${DSN_SENTRY}
endif

ifneq "${BUILD_ENV}" ""
	GO_LDFLAGS+=-X github.com/ProtonMail/proton-bridge/v3/internal/constants.BuildEnv=${BUILD_ENV}
endif

GO_LDFLAGS_LAUNCHER:=${GO_LDFLAGS}
ifeq "${TARGET_OS}" "windows"
	#GO_LDFLAGS+=-H=windowsgui # Disabled so we can inspect trace logs from the bridge for debugging.
	GO_LDFLAGS_LAUNCHER+=-H=windowsgui # Having this flag prevent a temporary cmd.exe window from popping when starting the application on Windows 11.
endif

BUILD_FLAGS+=-ldflags '${GO_LDFLAGS}'
BUILD_FLAGS_LAUNCHER+=-ldflags '${GO_LDFLAGS_LAUNCHER}'
DEPLOY_DIR:=cmd/${TARGET_CMD}/deploy
DIRNAME:=$(shell basename ${CURDIR})

LAUNCHER_EXE:=proton-bridge
BRIDGE_EXE=bridge
BRIDGE_GUI_EXE_NAME=bridge-gui
BRIDGE_GUI_EXE=${BRIDGE_GUI_EXE_NAME}
LAUNCHER_PATH:=cmd/launcher

ifeq "${TARGET_OS}" "windows"
	BRIDGE_EXE:=${BRIDGE_EXE}.exe
	BRIDGE_GUI_EXE:=${BRIDGE_GUI_EXE}.exe
	LAUNCHER_EXE:=${LAUNCHER_EXE}.exe
	RESOURCE_FILE:=resource.syso
endif
ifeq "${TARGET_OS}" "darwin"
	BRIDGE_EXE_NAME:=${BRIDGE_EXE}
	BRIDGE_EXE:=${BRIDGE_EXE}.app
	BRIDGE_GUI_EXE:=${BRIDGE_GUI_EXE}.app
	EXE_BINARY_DARWIN:=Contents/MacOS/${BRIDGE_GUI_EXE_NAME}
	EXE_TARGET_DARWIN:=${DEPLOY_DIR}/${TARGET_OS}/${LAUNCHER_EXE}.app
	DARWINAPP_CONTENTS:=${EXE_TARGET_DARWIN}/Contents
endif
EXE_TARGET:=${DEPLOY_DIR}/${TARGET_OS}/${BRIDGE_EXE}
EXE_GUI_TARGET:=${DEPLOY_DIR}/${TARGET_OS}/${BRIDGE_GUI_EXE}

TGZ_TARGET:=bridge_${TARGET_OS}_${REVISION}.tgz

ifdef QT_API
	VENDOR_TARGET:=prepare-vendor update-qt-docs
else
	VENDOR_TARGET=update-vendor
endif

build: build-gui

build-gui: ${TGZ_TARGET}

build-nogui: ${EXE_NAME} build-launcher
ifeq "${TARGET_OS}" "darwin"
	mv ${BRIDGE_EXE} ${BRIDGE_EXE_NAME}
endif

go-build=go build $(1) -o $(2) $(3)
go-build-finalize=${go-build}
ifeq "${GOOS}-$(shell uname -m)" "darwin-arm64"
	go-build-finalize= \
		MACOSX_DEPLOYMENT_TARGET=${MACOS_MIN_VERSION_ARM64} CGO_ENABLED=1 CGO_CFLAGS="-mmacosx-version-min=${MACOS_MIN_VERSION_ARM64}" GOARCH=arm64 $(call go-build,$(1),$(2)_arm,$(3)) && \
		MACOSX_DEPLOYMENT_TARGET=${MACOS_MIN_VERSION_AMD64} CGO_ENABLED=1 CGO_CFLAGS="-mmacosx-version-min=${MACOS_MIN_VERSION_AMD64}" GOARCH=amd64 $(call go-build,$(1),$(2)_amd,$(3)) && \
		lipo -create -output $(2) $(2)_arm $(2)_amd && rm -f $(2)_arm $(2)_amd
endif

ifeq "${GOOS}" "windows"
	go-build-finalize= \
		$(if $(4),cp "${ROOT_DIR}/${RESOURCE_FILE}" ${4}  &&,) \
		$(call go-build,$(1),$(2),$(3)) \
		$(if $(4), && rm -f ${4},)
endif

${EXE_NAME}: gofiles  ${RESOURCE_FILE}
	$(call go-build-finalize,${BUILD_FLAGS},"${LAUNCHER_EXE}","./cmd/${TARGET_CMD}/","${ROOT_DIR}/cmd/${TARGET_CMD}/${RESOURCE_FILE}")
	mv ${LAUNCHER_EXE} ${BRIDGE_EXE}

build-launcher: ${RESOURCE_FILE}
	$(call go-build-finalize,${BUILD_FLAGS_LAUNCHER},"${LAUNCHER_EXE}","${ROOT_DIR}/${LAUNCHER_PATH}/","${ROOT_DIR}/${LAUNCHER_PATH}/${RESOURCE_FILE}")

versioner:
	go build ${BUILD_FLAGS} -o versioner utils/versioner/main.go

vault-editor:
	$(call go-build-finalize,-tags=debug,"vault-editor","./utils/vault-editor/main.go")

bridge-rollout:
	$(call go-build-finalize,, "bridge-rollout","./utils/bridge-rollout/bridge-rollout.go")

hasher:
	go build -o hasher utils/hasher/main.go

${TGZ_TARGET}: ${DEPLOY_DIR}/${TARGET_OS}
	rm -f $@
	tar -czvf $@ -C ${DEPLOY_DIR}/${TARGET_OS} .

${DEPLOY_DIR}/linux: ${EXE_TARGET} build-launcher
	cp -pf ./dist/${SRC_SVG} ${DEPLOY_DIR}/linux/logo.svg
	cp -pf ./LICENSE ${DEPLOY_DIR}/linux/
	cp -pf ./Changelog.md ${DEPLOY_DIR}/linux/
	cp -pf ./dist/${EXE_NAME}.desktop ${DEPLOY_DIR}/linux/
	mv ${LAUNCHER_EXE} ${DEPLOY_DIR}/linux/

${DEPLOY_DIR}/darwin: ${EXE_TARGET} build-launcher
	mv ${EXE_GUI_TARGET} ${EXE_TARGET_DARWIN}
	mv ${EXE_TARGET} ${DARWINAPP_CONTENTS}/MacOS/${BRIDGE_EXE_NAME}
	perl -i -pe"s/>${BRIDGE_GUI_EXE_NAME}/>${LAUNCHER_EXE}/g" ${DARWINAPP_CONTENTS}/Info.plist
	cp ./dist/${SRC_ICNS} ${DARWINAPP_CONTENTS}/Resources/${SRC_ICNS}
	cp LICENSE ${DARWINAPP_CONTENTS}/Resources/
	rm -rf "${DARWINAPP_CONTENTS}/Frameworks/QtWebEngine.framework"
	rm -rf "${DARWINAPP_CONTENTS}/Frameworks/QtWebView.framework"
	rm -rf "${DARWINAPP_CONTENTS}/Frameworks/QtWebEngineCore.framework"
	mv ${LAUNCHER_EXE}  ${DARWINAPP_CONTENTS}/MacOS/${LAUNCHER_EXE}
	./utils/remove_non_relative_links_darwin.sh "${EXE_TARGET_DARWIN}/${EXE_BINARY_DARWIN}"

${DEPLOY_DIR}/windows: ${EXE_TARGET} build-launcher
	cp ./dist/${SRC_ICO} ${DEPLOY_DIR}/windows/logo.ico
	cp LICENSE ${DEPLOY_DIR}/windows/LICENSE.txt
	mv ${LAUNCHER_EXE} ${DEPLOY_DIR}/windows/$(notdir ${LAUNCHER_EXE})
	# plugins are installed in a plugins folder while needs to be near the exe
	cp -rf ${DEPLOY_DIR}/windows/plugins/* ${DEPLOY_DIR}/windows/.
	rm -rf ${DEPLOY_DIR}/windows/plugins

${EXE_TARGET}: check-build-essentials ${EXE_NAME}
	cd internal/frontend/bridge-gui/bridge-gui && \
		BRIDGE_APP_FULL_NAME="${APP_FULL_NAME}" \
		BRIDGE_VENDOR="${APP_VENDOR}" \
		BRIDGE_APP_VERSION=${APP_VERSION} \
		BRIDGE_REVISION=${REVISION} \
		BRIDGE_TAG=${TAG} \
		BRIDGE_DSN_SENTRY=${DSN_SENTRY} \
 		BRIDGE_BUILD_TIME=${BUILD_TIME} \
		BRIDGE_GUI_BUILD_CONFIG=Release \
		BRIDGE_BUILD_ENV=${BUILD_ENV} \
		BRIDGE_INSTALL_PATH="${ROOT_DIR}/${DEPLOY_DIR}/${GOOS}" \
		./build.sh install
	mv "${ROOT_DIR}/${BRIDGE_EXE}" "$(ROOT_DIR)/${EXE_TARGET}"

WINDRES_YEAR:=$(shell date +%Y)
APP_VERSION_COMMA:=$(shell echo "${APP_VERSION}" | sed -e 's/[^0-9,.]*//g' -e 's/\./,/g')
${RESOURCE_FILE}: ./dist/info.rc ./dist/${SRC_ICO} .FORCE
	rm -f ./*.syso
	windres --target=pe-x86-64 \
		-I ./internal/frontend/share/ \
		-D ICO_FILE=${SRC_ICO} \
		-D EXE_NAME="${EXE_NAME}" \
		-D FILE_VERSION="${APP_VERSION}" \
		-D ORIGINAL_FILE_NAME="${EXE}" \
		-D PRODUCT_VERSION="${APP_VERSION}" \
		-D FILE_VERSION_COMMA=${APP_VERSION_COMMA} \
		-D YEAR=${WINDRES_YEAR} \
		-o ./${RESOURCE_FILE} $<

## Dev dependencies
.PHONY: install-devel-tools install-linter install-go-mod-outdated install-git-hooks
LINTVER:="v1.59.1"
LINTSRC:="https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"

install-dev-dependencies: install-devel-tools install-linter install-go-mod-outdated

install-devel-tools: check-has-go
	go get -v github.com/golang/mock/gomock
	go get -v github.com/golang/mock/mockgen
	go get -v github.com/go-delve/delve

install-linter: check-has-go
	curl -sfL $(LINTSRC) | sh -s -- -b $(shell go env GOPATH)/bin $(LINTVER)

install-go-mod-outdated:
	which go-mod-outdated || go install github.com/psampaz/go-mod-outdated@latest

install-git-hooks:
	cp utils/githooks/* .git/hooks/
	chmod +x .git/hooks/*

## Checks, mocks and docs
.PHONY: check-has-go check-build-essentials add-license change-copyright-year test bench coverage mocks lint-license lint-golang lint updates doc release-notes
check-has-go:
	@which go || (echo "Install Go-lang!" && exit 1)
	go version


check_is_installed=if ! which $(1) > /dev/null; then echo "Please install $(1)"; exit 1; fi
check-build-essentials:
	@$(call check_is_installed,zip)
	@$(call check_is_installed,unzip)
	@$(call check_is_installed,tar)
	@$(call check_is_installed,curl)
ifneq "${GOOS}" "windows"
	@$(call check_is_installed,cmake)
	@$(call check_is_installed,ninja)
endif

add-license:
	./utils/missing_license.sh add

change-copyright-year:
	./utils/missing_license.sh change-year

GOCOVERAGE=-covermode=count -coverpkg=github.com/ProtonMail/proton-bridge/v3/internal/...,github.com/ProtonMail/proton-bridge/v3/pkg/...,
GOCOVERDIR=-args -test.gocoverdir=$$PWD/coverage

test: gofiles
	mkdir -p coverage/unit-${GOOS}
	go test \
		-v -timeout=20m -p=1 -count=1 \
		${GOCOVERAGE} \
		-run=${TESTRUN} ./internal/... ./pkg/... \
		${GOCOVERDIR}/unit-${GOOS}

test-race: gofiles
	go test -v -timeout=40m -p=1 -count=1 -race -failfast -run=${TESTRUN} ./internal/... ./pkg/...

test-integration: gofiles
	mkdir -p coverage/integration
	go test \
		-v -timeout=60m -p=1 -count=1 -tags=test_integration \
		${GOCOVERAGE} \
		github.com/ProtonMail/proton-bridge/v3/tests \
		${GOCOVERDIR}/integration


test-integration-debug: gofiles
	dlv test github.com/ProtonMail/proton-bridge/v3/tests -- -test.v -test.timeout=10m -test.parallel=1 -test.count=1

test-integration-race: gofiles
	go test -v -timeout=60m -p=1 -count=1 -race -failfast github.com/ProtonMail/proton-bridge/v3/tests

test-integration-nightly: gofiles
	mkdir -p coverage/integration
	gotestsum \
		--junitfile tests/result/feature-tests.xml -- \
		-v -timeout=90m -p=1 -count=1 -tags=test_integration \
		${GOCOVERAGE} \
		github.com/ProtonMail/proton-bridge/v3/tests \
		${GOCOVERDIR}/integration \
		nightly

fuzz: gofiles
	go test -fuzz=FuzzUnmarshal 	 -parallel=4 -fuzztime=60s $(PWD)/internal/legacy/credentials
	go test -fuzz=FuzzNewParser 	 -parallel=4 -fuzztime=60s $(PWD)/pkg/message/parser
	go test -fuzz=FuzzReadHeaderBody -parallel=4 -fuzztime=60s $(PWD)/pkg/message
	go test -fuzz=FuzzDecodeHeader 	 -parallel=4 -fuzztime=60s $(PWD)/pkg/mime
	go test -fuzz=FuzzDecodeCharset  -parallel=4 -fuzztime=60s $(PWD)/pkg/mime

bench:
	go test -run '^$$' -bench=. -memprofile bench_mem.pprof -cpuprofile bench_cpu.pprof ./internal/store
	go tool pprof -png -output bench_mem.png bench_mem.pprof
	go tool pprof -png -output bench_cpu.png bench_cpu.pprof

coverage: test
	go tool cover -html=/tmp/coverage.out -o=coverage.html

mocks:
	mockgen --package mocks github.com/ProtonMail/proton-bridge/v3/internal/bridge TLSReporter,ProxyController,Autostarter > tmp
	mv tmp internal/bridge/mocks/mocks.go
	mockgen --package mocks github.com/ProtonMail/gluon/async PanicHandler > internal/bridge/mocks/async_mocks.go
	mockgen --package mocks github.com/ProtonMail/gluon/reporter Reporter > internal/bridge/mocks/gluon_mocks.go
	mockgen --package mocks github.com/ProtonMail/proton-bridge/v3/internal/updater Downloader,Installer > internal/updater/mocks/mocks.go
	mockgen --package mocks github.com/ProtonMail/proton-bridge/v3/internal/telemetry HeartbeatManager > internal/telemetry/mocks/mocks.go
	cp internal/telemetry/mocks/mocks.go internal/bridge/mocks/telemetry_mocks.go
	mockgen --package mocks github.com/ProtonMail/proton-bridge/v3/internal/services/userevents \
EventSource,EventIDStore  > internal/services/userevents/mocks/mocks.go
	mockgen --package userevents github.com/ProtonMail/proton-bridge/v3/internal/services/userevents \
EventSubscriber,MessageEventHandler,LabelEventHandler,AddressEventHandler,RefreshEventHandler,UserEventHandler,UserUsedSpaceEventHandler > tmp
	mv tmp internal/services/userevents/mocks_test.go
	mockgen --package mocks github.com/ProtonMail/proton-bridge/v3/internal/events EventPublisher \
> internal/events/mocks/mocks.go
	mockgen --package mocks github.com/ProtonMail/proton-bridge/v3/internal/services/useridentity IdentityProvider,Telemetry \
> internal/services/useridentity/mocks/mocks.go
	mockgen --self_package "github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice" -package syncservice github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice \
ApplyStageInput,BuildStageInput,BuildStageOutput,DownloadStageInput,DownloadStageOutput,MetadataStageInput,MetadataStageOutput,\
StateProvider,Regulator,UpdateApplier,MessageBuilder,APIClient,Reporter,DownloadRateModifier \
> tmp
	mv tmp internal/services/syncservice/mocks_test.go
	mockgen --package mocks github.com/ProtonMail/gluon/connector IMAPStateWrite > internal/services/imapservice/mocks/mocks.go

lint: gofiles lint-golang lint-license lint-dependencies lint-changelog lint-bug-report

lint-license:
	./utils/missing_license.sh check

lint-dependencies:
	./utils/dependency_license.sh check

lint-changelog:
	./utils/changelog_linter.sh Changelog.md

lint-golang:
	which golangci-lint || $(MAKE) install-linter
	$(info linting with GOMAXPROCS=${GOMAXPROCS})
	golangci-lint run ./...

lint-bug-report:
	python3 utils/validate_bug_report_file.py --file "internal/frontend/bridge-gui/bridge-gui/qml/Resources/bug_report_flow.json"

lint-bug-report-preview:
	python3 utils/validate_bug_report_file.py --file "internal/frontend/bridge-gui/bridge-gui/qml/Resources/bug_report_flow.json" --preview

updates: install-go-mod-outdated
	# Uncomment the "-ci" to fail the job if something can be updated.
	go list -u -m -json all | go-mod-outdated -update -direct #-ci

doc:
	godoc -http=:6060

release-notes: release-notes/bridge_stable.html release-notes/bridge_early.html utils/release_notes.sh

release-notes/%.html: release-notes/%.md
	./utils/release_notes.sh $^

.PHONY: gofiles
# Following files are for the whole app so it makes sense to have them in bridge package.
# (Options like cmd or internal were considered and bridge package is the best place for them.)
gofiles: ./internal/bridge/credits.go
./internal/bridge/credits.go: ./utils/credits.sh go.mod
	cd ./utils/ && ./credits.sh bridge

## Run and debug
.PHONY: run run-qt run-qt-cli run-nogui run-cli run-noninteractive run-debug run-gui-tester clean-vendor clean-frontend-qt clean-frontend-qt-common clean

LOG?=debug
LOG_IMAP?=client # client/server/all, or empty to turn it off
LOG_SMTP?=--log-smtp # empty to turn it off
RUN_FLAGS?=-l=${LOG} --log-imap=${LOG_IMAP} ${LOG_SMTP}

run: run-qt

run-cli: run-nogui

run-noninteractive: build-nogui clean-vendor gofiles
	PROTONMAIL_ENV=dev ./${LAUNCHER_EXE} ${RUN_FLAGS} -n

run-qt: build-gui
ifeq "${TARGET_OS}" "darwin"
	PROTONMAIL_ENV=dev ${DARWINAPP_CONTENTS}/MacOS/${LAUNCHER_EXE}  ${RUN_FLAGS}
else
	PROTONMAIL_ENV=dev ./${DEPLOY_DIR}/${TARGET_OS}/${LAUNCHER_EXE} ${RUN_FLAGS}
endif

run-nogui: build-nogui clean-vendor gofiles
	PROTONMAIL_ENV=dev ./${LAUNCHER_EXE} ${RUN_FLAGS} -c

run-debug:
	dlv debug \
		--build-flags "-ldflags '-X github.com/ProtonMail/proton-bridge/v3/internal/constants.Version=3.1.0+git'" \
		./cmd/Desktop-Bridge/main.go \
		-- \
		-n -l=trace

ifeq "${TARGET_OS}" "windows"
	EXE_SUFFIX=.exe
endif

bridge-gui-tester:  build-gui
	cp ./cmd/Desktop-Bridge/deploy/${TARGET_OS}/bridge-gui${EXE_SUFFIX} .
	cd ./internal/frontend/bridge-gui/bridge-gui-tester && cmake . && make

run-gui-tester: bridge-gui-tester
	# copying tester as bridge so bridge-gui will start it and connect to it automatically
	cp ./internal/frontend/bridge-gui/bridge-gui-tester/bridge-gui-tester${EXE_SUFFIX} bridge${EXE_SUFFIX}
	./bridge-gui${EXE_SUFFIX}


clean-vendor:
	rm -rf ./vendor

clean-gui:
	cd internal/frontend/bridge-gui/ && \
		rm -f BuildConfig.h && \
		rm -rf cmake-build-*/

clean-vcpkg:
	git submodule deinit -f ./extern/vcpkg
	rm -rf ./.git/submodule/vcpkg
	rm -rf ./extern/vcpkg
	git checkout -- extern/vcpkg

clean: clean-vendor clean-gui clean-vcpkg
	rm -rf vendor-cache
	rm -rf cmd/Desktop-Bridge/deploy
	rm -rf cmd/Import-Export/deploy
	rm -f build last.log mem.pprof main.go
	rm -f ./*.syso
	rm -f release-notes/bridge.html
	rm -f release-notes/import-export.html
	rm -f ${LAUNCHER_EXE} ${BRIDGE_EXE} ${BRIDGE_EXE_NAME}


.PHONY: generate
generate:
	go generate ./...
	$(MAKE) build

.FORCE:

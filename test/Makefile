.PHONY: check-go check-godog install-godog test test-bridge test-live test-live-bridge test-stage test-debug test-live-debug bench

export GO111MODULE=on
export BRIDGE_VERSION:=2.3.0+integrationtests
export VERBOSITY?=fatal
export TEST_DATA=testdata

# Tests do not run in parallel. This will overrule user settings.
MAKEFLAGS=-j1

check-go:
	@which go || (echo "Install Go-lang!" && exit 1)
check-godog:
	@which godog || $(MAKE) install-godog
install-godog: check-go
	go get github.com/cucumber/godog/cmd/godog@v0.12.1

test: test-bridge
test-bridge: FEATURES ?= features
test-bridge: check-godog
	TEST_ENV=fake TEST_ACCOUNTS=accounts/fake.json godog --tags="~@ignore" $(FEATURES)

# Doesn't work in parallel!
# Provide TEST_ACCOUNTS with your accounts.
test-live: test-live-bridge test-live-ie
test-live-bridge: FEATURES ?= features
test-live-bridge: check-godog
	TEST_ENV=live godog --tags="~@ignore && ~@ignore-live" $(FEATURES)

# Doesn't work in parallel!
# Provide TEST_ACCOUNTS with your accounts.
# We need to pass build tag which is not possible with godog command.
# Tests against staging env are intended for debug purposes or checking new changes on API.
test-stage:
	TEST_ENV=live go test -tags=$(TAGS) -- $(FEATURES)

test-debug:
	TEST_ENV=fake TEST_ACCOUNTS=accounts/fake.json dlv test -- $(FEATURES)

test-live-debug:
	TEST_ENV=live dlv test -- $(FEATURES)

# -run flag is not working anyway, but lets keep it there to note we really do not want to run tests.
# To properly benchmark sync/fetch, we need everything empty. For that is better to start everything
# again and safest way is to run only one loop per run.
bench:
	TEST_DATA=../testdata go test -run='^$$' -bench=. -benchtime=1x -timeout=60m ./benchmarks/...

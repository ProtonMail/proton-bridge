.PHONY: check-has-go install-godog test test-live test-debug test-live-debug

export GO111MODULE=on
export BRIDGE_VERSION:=dev-integrationtests
export VERBOSITY?=fatal
export TEST_DATA=testdata

check-has-go:
	@which go || (echo "Install Go-lang!" && exit 1)

install-godog: check-has-go
	go get github.com/cucumber/godog/cmd/godog@v0.8.1

test:
	which godog || $(MAKE) install-godog
	TEST_ENV=fake TEST_ACCOUNTS=accounts/fake.json godog --tags="~@ignore" $(FEATURES)

# Doesn't work in parallel!
# Provide TEST_ACCOUNTS with your accounts.
test-live:
	which godog || $(MAKE) install-godog
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

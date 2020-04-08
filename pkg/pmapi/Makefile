export GO111MODULE=on

LINTVER="v1.21.0"
LINTSRC="https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"

check-has-go:
	@which go || (echo "Install Go-lang!" && exit 1)

install-dev-dependencies: install-linter

install-linter: check-has-go
	curl -sfL $(LINTSRC) | sh -s -- -b $(shell go env GOPATH)/bin $(LINTVER)

lint:
	which golangci-lint || $(MAKE) install-linter
	golangci-lint run ./...																						\

test:
	go test -run=${TESTRUN} ./...

TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=aptible
TEST_COUNT?=1
CUR_DIR = $(shell echo "${PWD}")

default: build

build: fmtcheck
	go install

gen:
	go generate ./...

test: fmtcheck
	go test $(TEST) $(TESTARGS) -timeout=120s -parallel=4

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v -count $(TEST_COUNT) -parallel 20 $(TESTARGS) -timeout 120m

fmt:
	gofmt -s -w .

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint:
	@./bin/golangci-lint run ./$(PKG_NAME)/...
	@docker run -v $(CUR_DIR):/src bflad/tfproviderlint:0.14.0 ./...

tools:
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.24.0
	@docker pull bflad/tfproviderlint:0.14.0

.PHONY: build gen test testacc fmt fmtcheck lint tools

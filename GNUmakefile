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
	@golangci-lint run ./$(PKG_NAME)/...
	@tfproviderlint ./...

tools:
	@go mod vendor
	@go install github.com/bflad/tfproviderlint/cmd/tfproviderlint
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: build gen test testacc fmt fmtcheck lint tools

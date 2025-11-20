TEST?=./...
TEST_PARALLEL?=12
TEST_TIMEOUT?=30m
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=aptible
TEST_COUNT?=1
CUR_DIR = $(shell echo "${PWD}")
LOCAL_TARGET=$(shell uname -s -m | tr '[:upper:]' '[:lower:]' | tr ' ' '_')
LOCAL_VERSION=0.0.0+local

default: build

build: fmtcheck
	go build

local-install:
	go build -o terraform-provider-aptible
	@mkdir -p "$$HOME/.terraform.d/plugins/aptible.com/aptible/aptible/$(LOCAL_VERSION)/$(LOCAL_TARGET)"
	@# If the file isn't explicitly deleted before the copy then terraform fails to load when changes are made
	@rm "$$HOME/.terraform.d/plugins/aptible.com/aptible/aptible/$(LOCAL_VERSION)/$(LOCAL_TARGET)/terraform-provider-aptible" || true
	@mv terraform-provider-aptible "$$HOME/.terraform.d/plugins/aptible.com/aptible/aptible/$(LOCAL_VERSION)/$(LOCAL_TARGET)"
	@echo "Installed as provider aptible.com/aptible/aptible version $(LOCAL_VERSION)"

gen:
	go generate ./...

test: fmtcheck
	go test $(TEST) $(TESTARGS) -timeout=120s -parallel=4

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v -count $(TEST_COUNT) -parallel $(TEST_PARALLEL) -timeout $(TEST_TIMEOUT) $(TESTARGS)

fmt:
	gofmt -s -w .

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint:
	@bin/golangci-lint run ./$(PKG_NAME)/...
	@$$(go env GOPATH)/bin/tfproviderlint ./...

tools:
	@go mod vendor
	@go install github.com/bflad/tfproviderlint/cmd/tfproviderlint
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v2.6.2

.PHONY: build gen test testacc fmt fmtcheck lint tools local-install

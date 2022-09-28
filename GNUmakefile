TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=aptible
TEST_COUNT?=1
CUR_DIR = $(shell echo "${PWD}")
TARGET=darwin_amd64
LOCAL_TARGET=darwin_$(shell uname -p)64

default: build

build: fmtcheck
	go build

build_local:
	TARGET=$(LOCAL_TARGET) go build

local-install: build_local
	@mkdir -p "$$HOME/.terraform.d/plugins/aptible.com/aptible/aptible/0.0.0+local/$(LOCAL_TARGET)"
	@# If the file isn't explicitly deleted before the copy then terraform fails to load when changes are made
	@rm "$$HOME/.terraform.d/plugins/aptible.com/aptible/aptible/0.0.0+local/$(LOCAL_TARGET)/terraform-provider-aptible" || true
	@cp terraform-provider-aptible "$$HOME/.terraform.d/plugins/aptible.com/aptible/aptible/0.0.0+local/$(LOCAL_TARGET)"
	@echo "Installed as provider aptible.com/aptible/aptible version 0.0.0+local"

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
	@bin/golangci-lint run ./$(PKG_NAME)/...
	@tfproviderlint ./...

tools:
	@go mod vendor
	@go install github.com/bflad/tfproviderlint/cmd/tfproviderlint
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.35.2

.PHONY: build gen test testacc fmt fmtcheck lint tools local-install


GO ?= go
GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)
OUTPUT_DIR := _output/$(GOOS)/$(GOARCH)

VERSION ?= $(shell git describe --dirty --always --tags | sed 's/-/./g')
GO_LDFLAGS := -ldflags '-X registry-cli/version.BuildVersion=$(VERSION)'

all: build
build: fmt vet output
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -buildmode=pie $(GO_LDFLAGS) -o '$(OUTPUT_DIR)/registrycli' ./cmd
output:
	mkdir -p "$(OUTPUT_DIR)"
fmt:
	$(GO) fmt ./...
vet:
	$(GO) vet ./...

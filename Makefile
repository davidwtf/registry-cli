
GO ?= go
OS ?= $(shell $(GO) env GOOS)
ARCH ?= $(shell $(GO) env GOARCH)

VERSION ?= $(shell git describe --dirty --always --tags | sed 's/-/./g')
GO_LDFLAGS := -ldflags '-X registry-cli/version.BuildVersion=$(VERSION)'
PLATFORMS := linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64

.PHONY: all
all: fmt vet build

.PHONY: build
build: go.build.$(OS)_$(ARCH)

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: build.all
build.all: fmt vet $(foreach p,$(PLATFORMS),$(addprefix go.build., $(p)))

.PHONY: go.build.%
go.build.%:
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval OUTPUT_DIR := _output/$(OS)/$(ARCH))
	mkdir -p "$(OUTPUT_DIR)"
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=pie $(GO_LDFLAGS) -o '$(OUTPUT_DIR)/registrycli' ./cmd

.PHONY: package
package: build.all tar.all


.PHONY: tar.all
tar.all: $(foreach p,$(PLATFORMS),$(addprefix tar., $(p)))

.PHONY: tar.%
tar.%:
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval OUTPUT_DIR := _output/$(OS)/$(ARCH))
	tar czvf _output/registrycli-$(OS)-$(ARCH).tgz -C $(OUTPUT_DIR) registrycli

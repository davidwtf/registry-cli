.PHONY: all binaries
.DEFAULT: all
all: binaries

binaries:
	mkdir -p ouput/$(shell go env GOOS)/$(shell go env GOARCH)/
 	go build -o ouput/$(shell go env GOOS)/$(shell go env GOARCH)/registrycli ./cmd/

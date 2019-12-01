EXECUTABLE=gcqlsh

WINDOWS_ZIP=$(EXECUTABLE)_windows_amd64.zip
LINUX_ZIP=$(EXECUTABLE)_linux_amd64.tar.gz
DARWIN_ZIP=$(EXECUTABLE)_darwin_amd64.tar.gz

WINDOWS=$(EXECUTABLE).exe
LINUX=$(EXECUTABLE)
DARWIN=$(EXECUTABLE)_darwin_amd64

VERSION=$(shell cat VERSION)

.PHONY: all clean distr

all: clean build ## Build and run tests

build: windows linux darwin ## Build binaries
	@echo version: $(VERSION)

windows: $(WINDOWS) ## Build for Windows

linux: $(LINUX) ## Build for Linux

darwin: $(DARWIN) ## Build for Darwin (macOS)

$(WINDOWS):
	rm -rf $(WINDOWS_ZIP) $(WINDOWS)
	env GOOS=windows GOARCH=amd64 go build -v -o $(WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)"  ./cmd/gcqlsh.go
	zip $(WINDOWS_ZIP) $(WINDOWS)
	rm -rf $(WINDOWS)
	shasum -a 256 $(WINDOWS_ZIP) > $(WINDOWS_ZIP).sha256.sum

$(LINUX):
	rm -rf $(LINUX_ZIP) $(LINUX)
	env GOOS=linux GOARCH=amd64 go build -v -o $(LINUX) -ldflags="-s -w -X main.version=$(VERSION)"  ./cmd/gcqlsh.go
	tar zcvf $(LINUX_ZIP) $(LINUX)
	rm -rf $(LINUX)
	shasum -a 256 $(LINUX_ZIP) > $(LINUX_ZIP).sha256.sum

$(DARWIN):
	rm -rf $(DARWIN_ZIP) $(DARWIN)
	env GOOS=darwin GOARCH=amd64 go build -v -o $(DARWIN) -ldflags="-s -w -X main.version=$(VERSION)"  ./cmd/gcqlsh.go
	tar zcvf $(DARWIN_ZIP) $(DARWIN)
	rm -rf $(DARWIN)
	shasum -a 256 $(DARWIN_ZIP) > $(DARWIN_ZIP).sha256.sum

clean: ## Remove previous build
	rm -f $(DARWIN_ZIP) $(DARWIN) $(LINUX_ZIP) $(LINUX) $(WINDOWS_ZIP) $(WINDOWS)

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
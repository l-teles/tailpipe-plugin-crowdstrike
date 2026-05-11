TAILPIPE_INSTALL_DIR ?= ~/.tailpipe
BUILD_TAGS = netgo

PLUGIN_DIR    = $(TAILPIPE_INSTALL_DIR)/plugins/hub.tailpipe.io/plugins/l-teles/crowdstrike@latest
PLUGIN_BINARY = $(PLUGIN_DIR)/tailpipe-plugin-crowdstrike.plugin
VERSION_JSON  = $(PLUGIN_DIR)/version.json
VERSIONS_JSON = $(TAILPIPE_INSTALL_DIR)/plugins/versions.json

.PHONY: install test vet tidy

install:
	mkdir -p $(PLUGIN_DIR)
	go build -o $(PLUGIN_BINARY) -tags "${BUILD_TAGS}" *.go
	$(PLUGIN_BINARY) metadata > $(VERSION_JSON)
	rm -f $(VERSIONS_JSON)

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

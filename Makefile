DIST := dist
EXECUTABLE := gorush

GO ?= go
DEPLOY_ACCOUNT := appleboy
DEPLOY_IMAGE := $(EXECUTABLE)
GOFMT ?= gofmt "-s"

TARGETS ?= linux darwin windows openbsd
ARCHS ?= amd64 386
GOFILES := $(shell find . -name "*.go" -type f)
TAGS ?= sqlite
LDFLAGS ?= -X 'main.Version=$(VERSION)'

ifneq ($(shell uname), Darwin)
	EXTLDFLAGS = -extldflags "-static" $(null)
else
	EXTLDFLAGS =
endif

ifneq ($(DRONE_TAG),)
	VERSION ?= $(DRONE_TAG)
else
	VERSION ?= $(shell git describe --tags --always || git rev-parse --short HEAD)
endif

.PHONY: all
all: build

init:
ifeq ($(ANDROID_API_KEY),)
	@echo "Missing ANDROID_API_KEY Parameter"
	@exit 1
endif
ifeq ($(ANDROID_TEST_TOKEN),)
	@echo "Missing ANDROID_TEST_TOKEN Parameter"
	@exit 1
endif
	@echo "Already set ANDROID_API_KEY and ANDROID_TEST_TOKEN globale variable."

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

vet:
	$(GO) vet ./...

embedmd:
	@hash embedmd > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/campoy/embedmd; \
	fi
	embedmd -d *.md

lint:
	@hash revive > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/mgechev/revive; \
	fi
	revive -config .revive.toml ./... || exit 1

.PHONY: install
install: $(GOFILES)
	$(GO) install -v -tags '$(TAGS)' -ldflags '$(EXTLDFLAGS)-s -w $(LDFLAGS)'
	@echo "\n==>\033[32m Installed gorush to ${GOPATH}/bin/gorush\033[m"

.PHONY: build
build: $(EXECUTABLE)

.PHONY: $(EXECUTABLE)
$(EXECUTABLE): $(GOFILES)
	$(GO) build -v -tags '$(TAGS)' -ldflags '$(EXTLDFLAGS)-s -w $(LDFLAGS)' -o release/$@

.PHONY: misspell-check
misspell-check:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -error $(GOFILES)

.PHONY: misspell
misspell:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -w $(GOFILES)

.PHONY: test
test: init fmt-check
	@$(GO) test -v -cover -tags $(TAGS) -coverprofile coverage.txt ./... && echo "\n==>\033[32m Ok\033[m\n" || exit 1

release: release-dirs release-build release-copy release-compress release-check

release-dirs:
	mkdir -p $(DIST)/binaries $(DIST)/release

release-build:
	@hash gox > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/mitchellh/gox; \
	fi
	gox -os="$(TARGETS)" -arch="$(ARCHS)" -tags="$(TAGS)" -ldflags="$(EXTLDFLAGS)-s -w $(LDFLAGS)" -output="$(DIST)/binaries/$(EXECUTABLE)-$(VERSION)-{{.OS}}-{{.Arch}}"

.PHONY: release-compress
release-compress:
	@hash gxz > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/ulikunitz/xz/cmd/gxz; \
	fi
	cd $(DIST)/release/; for file in `find . -type f -name "*"`; do echo "compressing $${file}" && gxz -k -9 $${file}; done;

release-copy:
	$(foreach file,$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*),cp $(file) $(DIST)/release/$(notdir $(file));)

release-check:
	cd $(DIST)/release; $(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

build_linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags '$(TAGS)' -ldflags '$(EXTLDFLAGS)-s -w $(LDFLAGS)' -o release/linux/amd64/$(DEPLOY_IMAGE)

build_linux_i386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -tags '$(TAGS)' -ldflags '$(EXTLDFLAGS)-s -w $(LDFLAGS)' -o release/linux/i386/$(DEPLOY_IMAGE)

build_linux_arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -tags '$(TAGS)' -ldflags '$(EXTLDFLAGS)-s -w $(LDFLAGS)' -o release/linux/arm64/$(DEPLOY_IMAGE)

build_linux_arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -tags '$(TAGS)' -ldflags '$(EXTLDFLAGS)-s -w $(LDFLAGS)' -o release/linux/arm/$(DEPLOY_IMAGE)

build_linux_lambda:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags 'lambda' -ldflags '$(EXTLDFLAGS)-s -w $(LDFLAGS)' -o release/linux/lambda/$(DEPLOY_IMAGE)

clean:
	$(GO) clean -modcache -x -i ./...
	find . -name coverage.txt -delete
	find . -name *.tar.gz -delete
	find . -name *.db -delete
	-rm -rf release dist .cover

rpc/example/node/gorush_*_pb.js: rpc/proto/gorush.proto
	@test -s ./node_modules/.bin/grpc_tools_node_protoc_plugin > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		npm install grpc-tools@1.8; \
	fi
	protoc -I vendor/github.com/golang -I rpc/proto rpc/proto/gorush.proto --js_out=import_style=commonjs,binary:rpc/example/node/ --grpc_out=rpc/example/node/ --plugin=protoc-gen-grpc=./node_modules/.bin/grpc_tools_node_protoc_plugin

rpc/proto/gorush.pb.go: rpc/proto/gorush.proto
	protoc -I vendor/github.com/golang -I rpc/proto rpc/proto/gorush.proto --go_out=plugins=grpc:rpc/proto

generate_proto: rpc/proto/gorush.pb.go rpc/example/node/gorush_*_pb.js

version:
	@echo $(VERSION)

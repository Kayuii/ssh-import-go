GOCMD:=$(shell which go)
GOLINT:=$(shell which golint)
GOIMPORT:=$(shell which goimports)
GOFMT:=$(shell which gofmt)
GOBUILD:= CGO_ENABLED=0 $(GOCMD) build -trimpath -ldflags '-X "github.com/kayuii/ssh-import-go/version.BinaryVersion=${BINARYVERSION}" \
		-X "github.com/kayuii/ssh-import-go/version.GoVersion=${GOVERSION}" \
		-X "github.com/kayuii/ssh-import-go/version.GitLastLog=${GITLASTLOG}" \
		-w -s -buildid='
GOINSTALL:=$(GOCMD) install
GOCLEAN:=$(GOCMD) clean
GOTEST:=$(GOCMD) test
GOGET:=$(GOCMD) get
GOLIST:=$(GOCMD) list
GOVET:=$(GOCMD) vet
GOPATH:=$(shell $(GOCMD) env GOPATH)
u := $(if $(update),-u)

GOVERSION=$(shell go version)
BINARYVERSION=$(shell git tag)
GITLASTLOG=$(shell git log --pretty=format:'%h - %s (%cd) <%an>' -1)
BINARY_NAME:=ssh-import
PACKAGES:=$(shell $(GOLIST) github.com/kayuii/ssh-import-go github.com/kayuii/ssh-import-go/cmd/ssh-import)
GOFILES:=$(shell find . -name "*.go" -type f)

export GO111MODULE := on

all: build

mini: test build-mini

.PHONY: build
build: deps
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/ssh-import/

.PHONY: build-mini
build-mini: deps
	$(GOBUILD) -ldflags "-s -w" -o $(BINARY_NAME)-mini ./cmd/ssh-import

.PHONY: install
install: deps
	$(GOINSTALL) ./cmd/ssh-import

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: deps
deps:
	$(GOGET) github.com/urfave/cli

.PHONY: devel-deps
devel-deps:
	GO111MODULE=off $(GOGET) -v -u \
		golang.org/x/lint/golint

.PHONY: lint
lint: devel-deps
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

.PHONY: vet
vet: deps devel-deps
	$(GOVET) $(PACKAGES)

.PHONY: fmt
fmt:
	$(GOFMT) -s -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -s -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

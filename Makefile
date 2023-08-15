SOURCE_DIRS = main.go cmd pkg
export GO111MODULE=on

Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)

LDFLAGS := "-s -w -X github.com/self-actuated/actuated-cli/pkg.Version=$(Version) -X github.com/self-actuated/actuated-cli/pkg.GitCommit=$(GitCommit)"

.PHONY: all
all: gofmt test build dist

.PHONY: build
build:
	go build

.PHONY: gofmt
gofmt:
	@test -z $(shell gofmt -l -s $(SOURCE_DIRS) ./ | tee /dev/stderr) || (echo "[WARN] Fix formatting issues with 'make gofmt'" && exit 1)

.PHONY: test
test:
	CGO_ENABLED=0 go test $(shell go list ./... | grep -v /vendor/|xargs echo) -cover

.PHONY: dist
dist:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS)  -o bin/actuated-cli
	CGO_ENABLED=0 GOOS=darwin go build -ldflags $(LDFLAGS)  -o bin/actuated-cli-darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -ldflags $(LDFLAGS) -installsuffix cgo -o bin/actuated-cli-darwin-arm64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS)  -o bin/actuated-cli.exe

.PHONY: all serverless deps docker docker-cgo clean docs test test-race test-integration fmt lint install deploy-docs build
DEST_DIR           = ./dist

CMD ?= go-socks5
VERSION = $(shell cat VERSION 2>/dev/null || echo '0.0.0')
VER_CUT   := $(shell echo $(VERSION) | cut -c2-)
GITREV = $(shell git rev-parse --short HEAD || echo unknown)
GITBRANCH = $(shell git branch | sed -n -e 's/^\* \(.*\)/\1/p')
BUILDTIME = $(shell date +'%Y-%m-%d_%T')
LDFLAGS = -ldflags "-w -s"
platform ?= amd64
gotool:
	@go mod tidy
	@go fmt ./...
	@#go vet ./...


all: clean gotool $(TARGET) $(CMD)

deps:
	@go mod tidy

clean:
	@rm -rf ./dist
	@rm -rf ./build
	@rm -rf ${CMD}.tar.xz

$(CMD):
	CGO_ENABLED=0 GOOS=linux GOARCH=$(platform)  go build ${LDFLAGS} -o build/$@/$@ ./cmd/$@

test: gotool
	go test -coverpkg=./... -coverprofile=cover.out -timeout 120s ./...
	go tool cover -html=cover.out -o coverage.html

pipeline-pack: all
	@if [ -e dist ] ; then rm -rf dist; fi
	@mkdir dist
	@$(foreach var,$(CMD),mkdir -p ./build/$(var)/conf;)
	@$(foreach var,$(CMD),cp -r ./cmd/$(var)/conf/ ./build/$(var)/;)
	@tar -C build -Jcf ${CMD}.tar.xz .
	@cp Dockerfile ./dist/Dockerfile
	sed -i 's/{{project}}/$(CMD)/g' ./dist/Dockerfile
	@mv ${CMD}.tar.xz ./dist

.PHONY: p
p: pipeline-pack


lint:
	golangci-lint run
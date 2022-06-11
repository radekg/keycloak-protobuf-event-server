BINARY      ?= keycloak-protobuf-event-server
GOARCH      ?= amd64
GOOS        ?= linux
VERSION     ?= $(shell git describe --tags --always)
BUILD_FLAGS ?=
LDFLAGS     ?= -X github.com/radekg/keycloak-protobuf-event-server/config.Version=$(VERSION) -w -s

.PHONY: genproto build docker.build test version

test:
	go test -v -count=1 `go list ./...`

build: build/$(BINARY)

build/$(BINARY):
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -o $(BINARY)-$(GOOS)-$(GOARCH) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" ./main.go

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY)-linux-amd64 $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" ./main.go

docker.build:
	docker build --build-arg GOOS=$(GOOS) --build-arg GOARCH=$(GOARCH) -t $(BINARY):latest -f Dockerfile .

version:
	@echo $(VERSION)

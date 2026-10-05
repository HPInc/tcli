IMAGE?=ghcr.io/hpinc/tcli
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
TCLI_LDFLAGS=-s -w -X github.com/hpinc/tcli/pkg/env.version=$(VERSION)
TRIVY_IMAGE:=ghcr.io/aquasecurity/trivy
BIN=bin/tcli.exe

ifndef OS
  OS=linux
  BIN=bin/tcli
endif

ifeq ($(OS),Windows_NT)
	OS=windows
	BIN=bin/tcli.exe
endif

ifndef ARCH
  ARCH=amd64
endif

all:
	TCLI_CONFIG_ROOT=tools \
		go run cmd/main.go $(MODULE) $(TAG)


build:
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) \
	go build -o $(BIN) \
	-ldflags "$(TCLI_LDFLAGS)" cmd/main.go

vendor:
	go mod vendor

vet:
	go vet ./...

imports:
	goimports -w .

tidy:
	go mod tidy

sanity: lint test trivy

lint:
	./tools/run_linter.sh

test:
	TCLI_INTEGRATION=1 go test ./...

docker:
	docker build --build-arg TCLI_VERSION=$(VERSION) \
	-t $(IMAGE):$(VERSION) -f tools/Dockerfile .

docker_run:
	docker run --rm $(IMAGE):$(VERSION)

trivy: docker
	docker run --rm \
	-v/var/run/docker.sock:/var/run/docker.sock \
	-v"$$HOME/Library/Caches:/root/.cache" \
	$(TRIVY_IMAGE) \
	image -q --severity HIGH,CRITICAL,MEDIUM,LOW --exit-code 1 $(IMAGE):$(VERSION)

clean:
	go clean

ci_clean: clean
	docker logout

install: build
	./tools/install.sh

.PHONY: all build vet imports tidy sanity lint gosec docker docker_run trivy clean install
.SILENT:

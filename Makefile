IMAGE:=tcli
TRIVY_IMAGE:=ghcr.io/aquasecurity/trivy
BIN=bin/tcli.exe

ifndef OS
  OS=linux
  BIN=bin/tcli
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
	-ldflags "-s -w" cmd/main.go

vet:
	go vet ./...

imports:
	goimports -w .

tidy:
	go mod tidy

sanity: lint trivy

lint:
	./tools/run_linter.sh

docker:
	docker build -t $(IMAGE) -f tools/Dockerfile .

docker_run:
	docker run --rm $(IMAGE)

trivy: docker
	docker run --rm \
	-v/var/run/docker.sock:/var/run/docker.sock \
	-v"$$HOME/Library/Caches:/root/.cache" \
	$(TRIVY_IMAGE) \
	image -q --severity HIGH,CRITICAL,MEDIUM,LOW --exit-code 1 $(IMAGE)

clean:
	go clean

install: build
	./tools/install.sh

.PHONY: all build vet imports tidy sanity lint gosec docker trivy clean install
.SILENT:

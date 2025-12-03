VERSION ?= "dev"
DATE := $(shell date -u +%Y-%m-%d_%H:%M:%S)
COMMIT := $(shell git rev-parse --short HEAD)

PROTO_DIR = api/v1
PROTO_FILE = $(PROTO_DIR)/shortener.proto
PROTO_OUT = pkg

.PHONY: build protoc

build:
	go build -ldflags="\
		-X 'main.buildVersion=$(VERSION)' \
		-X 'main.buildDate=$(DATE)' \
		-X 'main.buildCommit=$(COMMIT)'" \
	-o shortener cmd/shortener/main.go

protoc:
	mkdir -p $(PROTO_OUT)
	protoc --go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
	       --go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
	       $(PROTO_FILE)
VERSION ?= "dev"
DATE := $(shell date -u +%Y-%m-%d_%H:%M:%S)
COMMIT := $(shell git rev-parse --short HEAD)

PROTO_DIR = api/shortener/v1
PROTO_FILE = $(PROTO_DIR)/shortener.proto
PROTO_OUT = api/shortener/v1

.PHONY: build protoc

build:
	go build -ldflags="\
		-X 'main.buildVersion=$(VERSION)' \
		-X 'main.buildDate=$(DATE)' \
		-X 'main.buildCommit=$(COMMIT)'" \
	-o shortener cmd/shortener/main.go

protoc:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       $(PROTO_FILE)
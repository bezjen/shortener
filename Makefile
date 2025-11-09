VERSION ?= "dev"
DATE := $(shell date -u +%Y-%m-%d_%H:%M:%S)
COMMIT := $(shell git rev-parse --short HEAD)

.PHONY: build
build:
	go build -ldflags="\
		-X 'main.buildVersion=$(VERSION)' \
		-X 'main.buildDate=$(DATE)' \
		-X 'main.buildCommit=$(COMMIT)'" \
	-o shortener cmd/shortener/main.go
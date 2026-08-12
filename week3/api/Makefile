SHELL := /bin/bash

GO ?= go
APP := cloud3-api

.PHONY: help generate test run build fmt tidy check swagger clean hash-password

help:
	@echo "make generate  Generate Go code from openapi.yaml"
	@echo "make test      Run all Go tests"
	@echo "make run       Start the API locally"
	@echo "make build     Build the Go binary"
	@echo "make swagger   Start Swagger UI on port 8081"
	@echo "make hash-password  Generate a bcrypt password hash"
	@echo "make check     Format, generate, test and build"


# go tool oapi-codegen generate Go code bases on YAML
generate:
	@mkdir -p internal/api
	$(GO) tool oapi-codegen \
		-config oapi-codegen.yaml \
		openapi.yaml

test:
	$(GO) test ./...

run:
	$(GO) run .

build:
	@mkdir -p bin
	$(GO) build -o bin/$(APP) .

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

check: fmt generate test build

swagger:
	podman run --rm \
		--name cloud3-swagger \
		-p 8081:8080 \
		-e SWAGGER_JSON=/spec/openapi.yaml \
		-v "$(CURDIR)/openapi.yaml:/spec/openapi.yaml:ro" \
		docker.io/swaggerapi/swagger-ui

hash-password:
	$(GO) run ./cmd/hash-password

clean:
	rm -f bin/$(APP)
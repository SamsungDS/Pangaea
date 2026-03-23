# Copyright (c) 2025 Samsung Electronics and/or its Affiliates
.PHONY: help clean-go build-go generate-ofmf-service-client generate

APP_NAME := DCMFM
OPENAPI_REDFISH_YAML := api/redfish-openapi.yaml
GO_VERSION := 1.23.7
GO_INSTALL_LOCATION := /usr/local/go/bin/
GOFMT_OPTS := $(GO_INSTALL_LOCATION)gofmt -w /local
GENERATE_USER := $(shell id -u ${USER}):$(shell id -g ${USER})

help:
	@echo ""
	@echo "-------------------------------------------------------------------------------------------------------------------------"
	@echo "make clean-go                         - Remove all Go executables"
	@echo "make build-go                         - Build local $(APP_NAME) executables"
	@echo "make generate                         - Generate supporting code for the whole suite using openapi-generator-cli"
	@echo "make generate-ofmf-service-client     - Generate go client code for the redfish specification using openapi-generator-cli"
	@echo ""

clean-go:
	@echo "Clean up..."
	go clean
	rm -f $(APP_NAME)

build-go: clean-go
	@echo "Building DCMFM Go executables..."
	go build -o $(APP_NAME) ./$(APP_NAME).go
	ls -lh $(APP_NAME)

generate-ofmf-service-client:
	@echo "Clean up redfish client generated code (service)"
	rm -rf ./pkg/ofmf-service-client/
	@echo "Generating $(OPENAPI_REDFISH_YAML) go client using openapi-generator-cli"
	docker run -u $(GENERATE_USER) --rm -v ${PWD}:/local openapitools/openapi-generator-cli:v7.0.0 generate -i /local/$(OPENAPI_REDFISH_YAML) -g go -o /local/pkg/ofmf-service-client -p isGoSubmodule=true,withGoMod=false --package-name ofmf_service_client --additional-properties=sourceFolder=,outputAsLibrary=true,enumClassPrefix=true --ignore-file-override /local/api/ignore/.openapi-generator-ignore-client -t /local/api/templates/go
	# workaround for withGoMod=false not functioning with openapi-generator
	rm pkg/ofmf-service-client/go.mod
	rm pkg/ofmf-service-client/go.sum
	rm -rf pkg/ofmf-service-client/test # not currently used and have conflict with generated path (github.com/GIT_USER_ID/GIT_REPO_ID/client)

	@echo "Format files after generation to conform to project standard"
	docker run --rm -v ${PWD}:/local golang:$(GO_VERSION) $(GOFMT_OPTS)

generate: generate-ofmf-service-client

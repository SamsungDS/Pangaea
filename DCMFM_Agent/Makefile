# Copyright (c) 2025 Samsung Electronics and/or its Affiliates
.PHONY: help clean-go build-go generate-ofmf-service-client generate

APP_NAME := DCMFM_Agent
GO_VERSION := 1.23.7
GO_INSTALL_LOCATION := /usr/local/go/bin/
GOFMT_OPTS := $(GO_INSTALL_LOCATION)gofmt -w /local
GENERATE_USER := $(shell id -u ${USER}):$(shell id -g ${USER})

help:
	@echo ""
	@echo "-------------------------------------------------------------------------------------------------------------------------"
	@echo "make clean-go                         - Remove all Go executables"
	@echo "make build-go                         - Build local $(APP_NAME) executables"
	@echo ""

clean-go:
	@echo "Clean up..."
	go clean
	rm -f $(APP_NAME)

build-go: clean-go
	@echo "Building DCMFM Go executables..."
	go build -o $(APP_NAME) ./$(APP_NAME).go
	ls -lh $(APP_NAME)

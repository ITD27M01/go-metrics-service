.ONESHELL:
SHELL = /bin/bash
.PHONY: build clean update compile update clean get build clean-cache tidy

SERVER_BINNAME=server
AGENT_BINNAME=agent

# Go related variables.
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

SERVER_SOURCE=$(GOBASE)/cmd/server
AGENT_SOURCE=$(GOBASE)/cmd/agent

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## compile: Compile the binary.
build:
	$(MAKE) -s compile

## clean: Clean build files. Runs `go clean` internally.
clean:
	$(MAKE) go-clean

## update: Update modules
update:
	$(MAKE) go-update

compile: go-clean go-get-agent go-get-server build-agent build-server

go-update: go-clean-cache go-tidy

go-clean:
	@echo "  >  Cleaning build cache"
	@GOBIN=$(GOBIN) go clean
	@rm -rf $(GOBIN)

go-get-server:
	@echo "  >  Checking if there is any missing dependencies..."
	@cd $(SERVER_SOURCE); GOBIN=$(GOBIN) go get $(get)

go-get-agent:
	@echo "  >  Checking if there is any missing dependencies..."
	@cd $(AGENT_SOURCE); GOBIN=$(GOBIN) go get $(get)

build-server:
	@echo "  >  Building server binaries..."
	@cd $(SERVER_SOURCE); go build -o $(GOBIN)/$(SERVER_BINNAME) $(GOFILES)

build-agent:
	@echo "  >  Building agent binaries..."
	@cd $(AGENT_SOURCE); go build -o $(GOBIN)/$(AGENT_BINNAME) $(GOFILES)

go-clean-cache:
	@echo "  >  Clean modules cache..."
	@go clean -modcache

go-tidy:
	@echo "  >  Update modules..."
	@go mod tidy

go-test:
	@echo "  >  Test project..."
	@go test ./...

go-vet:
	@echo "  >  Vet project..."
	@go vet ./...
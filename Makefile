.ONESHELL:
SHELL = /bin/bash
.PHONY: build clean update compile update clean get build clean-cache tidy

SERVER_BINNAME=server
AGENT_BINNAME=agent
STATICLINT_BINNAME=staticlint

# Go related variables.
GOFLAGS += '-ldflags="-s -w"'
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

SERVER_SOURCE=$(GOBASE)/cmd/server
AGENT_SOURCE=$(GOBASE)/cmd/agent
STATICLINT_SOURCE=$(GOBASE)/cmd/staticlint

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

test: go-test go-statictest go-vet

compile: go-clean go-get-agent go-get-server build-agent build-server build-staticlint

go-update: go-clean go-clean-cache go-tidy go-download

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

go-get-staticlint:
	@echo "  >  Checking if there is any missing dependencies..."
	@cd $(STATICLINT_SOURCE); GOBIN=$(GOBIN) go get $(get)

build-server:
	@echo "  >  Building server binaries..."
	@cd $(SERVER_SOURCE); go build -o $(GOBIN)/$(SERVER_BINNAME) $(GOFILES)

build-agent:
	@echo "  >  Building agent binaries..."
	@cd $(AGENT_SOURCE); go build -o $(GOBIN)/$(AGENT_BINNAME) $(GOFILES)

build-staticlint:
	@echo "  >  Building staticlint binaries..."
	@cd $(STATICLINT_SOURCE); go build -o $(GOBIN)/$(STATICLINT_BINNAME) $(GOFILES)

go-clean-cache:
	@echo "  >  Clean modules cache..."
	@go clean -modcache

go-tidy:
	@echo "  >  Update modules..."
	@go mod tidy

go-download:
	@echo "  >  Download modules..."
	@go mod download

go-test:
	@echo "  >  Test project..."
	@go test ./...

go-container:
	@podman build -q -t go-metrics-server .
	@podman run --rm --name go-metrics-server go-metrics-server

go-statictest: go-container
	@echo " > Static test project..."
	@podman logs go-metrics-server

go-vet:
	@echo "  >  Vet project..."
	@go vet ./...

go-migrate:
	@echo "  >  Install migration tools..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

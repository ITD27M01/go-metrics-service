.ONESHELL:
SHELL = /bin/bash
.PHONY: build clean update compile update clean get build clean-cache tidy

BUILD_VERSION=2022.06
BUILD_DATE := $(shell date +"%Y%m%d%H%M")
BUILD_COMMIT := $(shell git log -n 1 --pretty=format:"%H")

SERVER_BINNAME=server
AGENT_BINNAME=agent
STATICLINT_BINNAME=staticlint

# Go related variables.
LDFLAGS := "-w -s -X main.buildVersion=${BUILD_VERSION} -X main.buildDate=${BUILD_DATE} -X main.buildCommit=${BUILD_COMMIT}"
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

SERVER_SOURCE=$(GOBASE)/cmd/server
AGENT_SOURCE=$(GOBASE)/cmd/agent
STATICLINT_SOURCE=$(GOBASE)/cmd/staticlint

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# For tests
DATABASE_DSN = "postgres://dbuser:dbpass@localhost:5432/metrics?sslmode=disable"

## compile: Compile the binary.
build:
	$(MAKE) -s compile

## compile: Compile the binaries for github
build-github:
	$(MAKE) -s compile-github

## clean: Clean build files. Runs `go clean` internally.
clean:
	$(MAKE) go-clean

## update: Update modules
update:
	$(MAKE) go-update

migration-tools:
	@echo "  >  Install migration tools..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate:
	@echo "  >  Do DB migrations..."
	@migrate -path db/migrations -database ${DATABASE_DSN} up

keys:
	@echo "	 > Build RSA keys for agent to server encryption"
	@openssl genrsa -out private-key.pem 4096
	@openssl rsa -in private-key.pem -outform PEM -pubout -out public-key.pem

test: go-test go-statictest go-vet

compile: go-clean go-get-agent go-get-server build-agent build-server build-staticlint

compile-github: go-clean go-get-agent go-get-server build-agent-github build-server-github

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
	@cd $(SERVER_SOURCE); go build -ldflags $(LDFLAGS) -o $(GOBIN)/$(SERVER_BINNAME) $(GOFILES)

build-server-github:
	@echo "  >  Building server binaries..."
	@cd $(SERVER_SOURCE); go build -ldflags $(LDFLAGS) -o $(SERVER_SOURCE)/$(SERVER_BINNAME) $(GOFILES)

build-agent:
	@echo "  >  Building agent binaries..."
	@cd $(AGENT_SOURCE); go build -ldflags $(LDFLAGS) -o $(GOBIN)/$(AGENT_BINNAME) $(GOFILES)

build-agent-github:
	@echo "  >  Building agent binaries..."
	@cd $(AGENT_SOURCE); go build -ldflags $(LDFLAGS) -o $(AGENT_SOURCE)/$(AGENT_BINNAME) $(GOFILES)

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

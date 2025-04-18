# Variables
PROTO_DIR := api/proto
GO_OUT_DIR := pkg/grpc
SERVICES := user product order telehealth cms notification

# Load environment variables from .env file
include .env
export

GOOSE_DRIVER ?= postgres
GOOSE_MIGRATION_DIR ?= internal/database/migrations


# Check for protoc installation
PROTOC := $(shell command -v protoc 2> /dev/null)

# Ensure the output directory exists
$(shell mkdir -p $(GO_OUT_DIR))

# Default target
proto-all: pfmt generate

# Generate Go code from all .proto files
generate: $(SERVICES)

# Rule to generate Go code for each service
$(SERVICES):
	@if [ -d $(PROTO_DIR)/$@ ]; then \
		echo "Generating Go code for $@ service..."; \
		OUT_DIR=$(GO_OUT_DIR)/$@; \
		mkdir -p $$OUT_DIR; \
		PROTO_FILES=$$(find $(PROTO_DIR)/$@ -name '*.proto'); \
		if [ -n "$$PROTO_FILES" ]; then \
			$(PROTOC) --proto_path=$(PROTO_DIR)/$@ \
				--go_out=$$OUT_DIR --go_opt=paths=source_relative \
				--go-grpc_out=$$OUT_DIR --go-grpc_opt=paths=source_relative \
				$$PROTO_FILES; \
		else \
			echo "Skipping $@ service: No .proto files found in $(PROTO_DIR)/$@."; \
		fi \
	else \
		echo "Skipping $@ service: $(PROTO_DIR)/$@ directory does not exist."; \
	fi

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -rf $(GO_OUT_DIR)/*

# Install necessary protoc plugins
install-plugins:
	@echo "Installing protoc-gen-go and protoc-gen-go-grpc..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Check if protoc is installed
check-protoc:
ifndef PROTOC
	$(error "protoc is not installed. Please install Protocol Buffers compiler.")
endif

# Prepare for build by generating protobuf files
prepare: check-protoc generate

# Build target (add your actual build commands here)
build: prepare
	@echo "Building the project..."
	# Add your go build commands here

.PHONY: all generate $(SERVICES) clean install-plugins check-protoc help prepare build

# Help target
help:
	@echo "Makefile for Chemist.ke"
	@echo ""
	@echo "Targets:"
	@echo "  generate     : Generate Go code from all .proto files"
	@echo "  clean        : Remove all generated files"
	@echo "  install-plugins : Install necessary protoc plugins"
	@echo "  check-protoc : Check if protoc is installed"
	# Database migration
	@echo "  create NAME=<name>    Create a new migration with the given name"
	@echo "  up                    Apply all available migrations"
	@echo "  down                  Roll back the last migration"
	@echo "  status                Print the status of all migrations"
	@echo "  help                  Show this help message"
	@echo "  help         : Show this help message"



# make sure the GOOSE_DBSTRING , GOOSE_DRIVER and GOOSE_MIGRATION_DIR are set
ifndef GOOSE_DBSTRING
$(error GOOSE_DBSTRING is not set)
endif

ifndef GOOSE_DRIVER
$(error GOOSE_DRIVER is not set)
endif

ifndef GOOSE_MIGRATION_DIR
$(error GOOSE_MIGRATION_DIR is not set)
endif

ifeq ($(NO_COLOR), 1)
COLOR_FLAG=--no-color
else
COLOR_FLAG=
endif

.PHONY: db-create
db-create:
	@read -p "Enter migration name: " NAME; \
	goose create $$NAME sql --dir $(GOOSE_MIGRATION_DIR)

.PHONY: db-up
db-up:
	@echo $(GOOSE_DBSTRING)
	goose up --dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)"

.PHONY: db-down
db-down:
	goose down --dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)"

.PHONY: db-status
db-status:
	goose status --dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)"

.PHONY: pfmt
# Format the proto files using buf format command pipe the result to the specific file
pfmt:
	buf format  $(PROTO_DIR)  -w


run-client:
	swag init -g cmd/api-gateway/gateway.go --parseDependency --output ./docs 
	go run cmd/api-gateway/gateway.go

#!/bin/bash

# Set the root directory
ROOT_DIR="."

# Function to create directories
create_dirs() {
    local parent=$1
    shift
    for dir in "$@"; do
        mkdir -p "$ROOT_DIR/$parent/$dir"
    done
}

# Create main directories
create_dirs "" cmd internal pkg api scripts deployments test docs web

# Create subdirectories
create_dirs "cmd" api-gateway user-service product-service order-service telehealth-service cms-service notification-service
create_dirs "internal" auth database logger middleware models
create_dirs "pkg" cache config errors utils
create_dirs "api" proto swagger
create_dirs "api/proto" user product order telehealth cms notification
create_dirs "deployments" docker kubernetes
#  where grpc service clients will be written
create_dirs "pkg" client
create_dirs "test" integration load
create_dirs "web" admin

# Create empty files
touch $ROOT_DIR/.gitignore
touch $ROOT_DIR/go.mod
touch $ROOT_DIR/go.sum
touch $ROOT_DIR/README.md

echo "Folder structure for $ROOT_DIR has been created successfully!"

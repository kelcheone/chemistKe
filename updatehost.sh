#!/bin/bash

# Check if both current and remote host are provided
if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <current_host> <remote_host>"
  exit 1
fi

CURRENT_HOST=$1
REMOTE_HOST=$2

# Files to update
GO_FILE="cmd/api-gateway/gateway.go"
JSON_FILE="docs/swagger.json"
YAML_FILE="docs/swagger.yaml"

# Ensure all files exist
for FILE in "$GO_FILE" "$JSON_FILE" "$YAML_FILE"; do
  if [ ! -f "$FILE" ]; then
    echo "File not found: $FILE"
    exit 1
  fi
done

# Replace in Go file
sed -i "s|// @host $CURRENT_HOST|// @host $REMOTE_HOST|" "$GO_FILE"
echo "Updated Go file: $GO_FILE"

# Replace in JSON file
sed -i "s|\"host\": \"$CURRENT_HOST\"|\"host\": \"$REMOTE_HOST\"|" "$JSON_FILE"
echo "Updated JSON file: $JSON_FILE"

# Replace in YAML file
sed -i "s|host: $CURRENT_HOST|host: $REMOTE_HOST|" "$YAML_FILE"
echo "Updated YAML file: $YAML_FILE"

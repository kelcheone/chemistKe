#!/bin/sh
set -e

# Check if .env file exists, create a minimal one if it doesn't
if [ ! -f .env ]; then
  echo "Warning: .env file not found, creating minimal configuration"
  echo "DB_URL=${DB_URL:-postgres://pg:chemistke@database:5432/chemist_ke}" > .env
  echo "GOOSE_DBSTRING=${GOOSE_DBSTRING:-postgres://pg:chemistke@database:5432/chemist_ke}" >> .env
fi

# Start services
./services &
./gateway

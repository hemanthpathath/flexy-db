#!/bin/bash

# Quick start script for flex-db

set -e

echo "🚀 Starting flex-db..."

# Get the script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GO_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"
PROJECT_ROOT="$( cd "$GO_DIR/.." && pwd )"

# Load environment variables from .env.local (if exists), otherwise from .env
if [ -f "$PROJECT_ROOT/.env.local" ]; then
    echo "📝 Loading environment variables from .env.local..."
    export $(cat "$PROJECT_ROOT/.env.local" | grep -v '^#' | xargs)
elif [ -f "$PROJECT_ROOT/.env" ]; then
    echo "📝 Loading environment variables from .env..."
    export $(cat "$PROJECT_ROOT/.env" | grep -v '^#' | xargs)
else
    echo "📝 No .env.local or .env file found. Using defaults."
    echo "💡 Tip: Copy .env.example to .env.local and customize it for your local setup"
fi

# Check if PostgreSQL is running (Docker)
if docker ps | grep -q flex-db-postgres; then
    echo "✅ PostgreSQL container is running"
elif docker ps -a | grep -q flex-db-postgres; then
    echo "🔄 Starting PostgreSQL container..."
    docker-compose -f "$PROJECT_ROOT/docker-compose.yml" up -d postgres
    echo "⏳ Waiting for PostgreSQL to be ready..."
    sleep 5
else
    echo "⚠️  PostgreSQL container not found. Starting with docker-compose..."
    docker-compose -f "$PROJECT_ROOT/docker-compose.yml" up -d postgres
    echo "⏳ Waiting for PostgreSQL to be ready..."
    sleep 5
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+"
    exit 1
fi

# Change to Go directory
cd "$GO_DIR"

# Download dependencies
echo "📦 Downloading dependencies..."
go mod download

# Run the server
echo "🎯 Starting gRPC server on port ${GRPC_PORT}..."
echo ""
go run ./cmd/dbaas-server


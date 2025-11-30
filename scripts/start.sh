#!/bin/bash

# flex-db Python Backend Start Script
# This script sets up and starts the Python backend server

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

echo "=== flex-db Python Backend ==="
echo ""

# Load environment variables from .env.local if it exists
if [ -f ".env.local" ]; then
    echo "Loading environment from .env.local..."
    set -a
    source .env.local
    set +a
fi

# Check if virtual environment exists, create if not
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
echo "Activating virtual environment..."
source venv/bin/activate

# Install dependencies
echo "Installing dependencies..."
pip install -r requirements.txt -q

# Check if PostgreSQL is running (assuming Docker setup)
if command -v docker &> /dev/null; then
    if ! docker ps --format '{{.Names}}' | grep -q "^flex-db-python-postgres$"; then
        echo "Starting PostgreSQL container..."
        cd "$PROJECT_DIR"
        if docker-compose up -d postgres 2>/dev/null || docker compose up -d postgres 2>/dev/null; then
            echo "Waiting for PostgreSQL to be ready..."
            sleep 5
            
            # Wait for PostgreSQL to be healthy
            max_attempts=30
            attempt=0
            while [ $attempt -lt $max_attempts ]; do
                if docker exec flex-db-python-postgres pg_isready -U postgres &> /dev/null; then
                    echo "PostgreSQL is ready"
                    break
                fi
                attempt=$((attempt + 1))
                sleep 1
            done
        fi
    fi
fi

# Run the server
echo ""
echo "Starting JSON-RPC server..."
python main.py

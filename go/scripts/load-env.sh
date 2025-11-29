#!/bin/bash

# Helper script to load environment variables from .env.local or .env
# Usage: source scripts/load-env.sh

# Get the script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GO_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"
PROJECT_ROOT="$( cd "$GO_DIR/.." && pwd )"

if [ -f "$PROJECT_ROOT/.env.local" ]; then
    echo "📝 Loading environment variables from .env.local..."
    export $(cat "$PROJECT_ROOT/.env.local" | grep -v '^#' | xargs)
elif [ -f "$PROJECT_ROOT/.env" ]; then
    echo "📝 Loading environment variables from .env..."
    export $(cat "$PROJECT_ROOT/.env" | grep -v '^#' | xargs)
else
    echo "⚠️  No .env.local or .env file found. Using defaults."
fi


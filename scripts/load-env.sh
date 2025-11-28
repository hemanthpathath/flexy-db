#!/bin/bash

# Helper script to load environment variables from .env.local or .env
# Usage: source scripts/load-env.sh

if [ -f .env.local ]; then
    echo "📝 Loading environment variables from .env.local..."
    export $(cat .env.local | grep -v '^#' | xargs)
elif [ -f .env ]; then
    echo "📝 Loading environment variables from .env..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "⚠️  No .env.local or .env file found. Using defaults."
fi


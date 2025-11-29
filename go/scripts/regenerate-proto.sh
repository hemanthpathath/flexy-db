#!/bin/bash

# Script to regenerate protobuf files
# Usage: ./scripts/regenerate-proto.sh

set -e

echo "🔄 Regenerating protobuf files..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "❌ protoc is not installed. Installing via brew..."
    brew install protobuf
fi

# Check if Go plugins are installed
export PATH=$PATH:$(go env GOPATH)/bin

if ! command -v protoc-gen-go &> /dev/null; then
    echo "📦 Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "📦 Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Regenerate proto files
echo "🔨 Generating Go code from proto files..."
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/dbaas.proto

echo "✅ Proto files regenerated successfully!"
echo ""
echo "Generated files:"
ls -lh api/proto/*.pb.go


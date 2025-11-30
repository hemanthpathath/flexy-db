#!/bin/bash

# flex-db Local Setup Script
# This script sets up the local development environment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
POSTGRES_CONTAINER="flex-db-postgres"

cd "$PROJECT_DIR"

echo "=========================================="
echo "  flex-db Local Setup Script"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "  ${YELLOW}→${NC} $1"
}

# Check prerequisites
echo "Step 1: Checking prerequisites..."
echo ""

# Check Python
if command -v python3 &> /dev/null; then
    PYTHON_VERSION=$(python3 --version | cut -d' ' -f2)
    print_success "Python 3 found: $PYTHON_VERSION"
else
    print_error "Python 3 is not installed. Please install Python 3.9+"
    exit 1
fi

# Check Docker
if command -v docker &> /dev/null; then
    print_success "Docker found"
else
    print_error "Docker is not installed. Please install Docker"
    exit 1
fi

# Check Docker Compose
if docker compose version &> /dev/null; then
    print_success "Docker Compose found"
else
    print_error "Docker Compose is not installed. Please install Docker Compose"
    exit 1
fi

echo ""

# Step 2: Create virtual environment
echo "Step 2: Setting up Python virtual environment..."
echo ""

if [ ! -d "venv" ]; then
    print_info "Creating virtual environment..."
    python3 -m venv venv
    print_success "Virtual environment created"
else
    print_info "Virtual environment already exists"
fi

# Activate virtual environment
print_info "Activating virtual environment..."
source venv/bin/activate

echo ""

# Step 3: Install dependencies
echo "Step 3: Installing Python dependencies..."
echo ""

if [ -f "requirements.txt" ]; then
    print_info "Installing packages from requirements.txt..."
    pip install --upgrade pip -q
    pip install -r requirements.txt -q
    print_success "Dependencies installed"
else
    print_error "requirements.txt not found"
    exit 1
fi

echo ""

# Step 4: Create .env.local file
echo "Step 4: Creating environment configuration..."
echo ""

ENV_FILE=".env.local"

if [ -f "$ENV_FILE" ]; then
    print_warning ".env.local already exists"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Keeping existing .env.local"
    else
        create_env_file=true
    fi
else
    create_env_file=true
fi

if [ "$create_env_file" = true ]; then
    cat > "$ENV_FILE" << 'EOF'
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres

# Control Database (for tenant metadata)
DB_CONTROL_NAME=dbaas_control

# Tenant Database Prefix
DB_TENANT_PREFIX=dbaas_tenant_

# Legacy (kept for compatibility)
DB_NAME=dbaas

# SSL Mode
DB_SSL_MODE=disable

# Server Configuration
JSONRPC_HOST=0.0.0.0
JSONRPC_PORT=5000

# Development
RELOAD=false
EOF
    print_success ".env.local created"
else
    print_info "Using existing .env.local"
fi

echo ""

# Step 5: Start PostgreSQL
echo "Step 5: Starting PostgreSQL database..."
echo ""

# Check if PostgreSQL container is already running
if docker ps --format '{{.Names}}' | grep -q "^${POSTGRES_CONTAINER}$"; then
    print_info "PostgreSQL container is already running"
else
    print_info "Starting PostgreSQL container..."
    if docker compose --profile dev up -d postgres 2>/dev/null; then
        print_success "PostgreSQL container started"
        print_info "Waiting for PostgreSQL to be ready..."
        sleep 5
        
        # Wait for PostgreSQL to be healthy
        max_attempts=30
        attempt=0
        while [ $attempt -lt $max_attempts ]; do
            if docker exec "${POSTGRES_CONTAINER}" pg_isready -U postgres &> /dev/null; then
                print_success "PostgreSQL is ready"
                break
            fi
            attempt=$((attempt + 1))
            sleep 1
        done
        
        if [ $attempt -eq $max_attempts ]; then
            print_error "PostgreSQL failed to become ready"
            exit 1
        fi
    else
        print_error "Failed to start PostgreSQL container"
        exit 1
    fi
fi

echo ""

# Step 6: Verify setup
echo "Step 6: Verifying setup..."
echo ""

# Check if we can import the app
print_info "Testing Python imports..."
if python3 -c "from app.config import config_from_env; print('OK')" 2>/dev/null; then
    print_success "Python imports working"
else
    print_error "Python imports failed"
    exit 1
fi

echo ""

# Summary
echo "=========================================="
echo "  Setup Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Start the server:"
echo "   source venv/bin/activate"
echo "   python main.py"
echo ""
echo "2. Or use the start script:"
echo "   ./scripts/start.sh"
echo ""
echo "3. Test the API:"
echo "   ./scripts/test_basic_operations.sh"
echo ""
echo "4. View the API documentation:"
echo "   curl http://localhost:5000/openrpc.json | jq"
echo ""
echo "The server will be available at:"
echo "  - JSON-RPC: http://localhost:5000/jsonrpc"
echo "  - Health:   http://localhost:5000/health"
echo ""
echo "PostgreSQL is running in Docker:"
echo "  - Host: localhost"
echo "  - Port: 5432"
echo "  - User: postgres"
echo "  - Password: postgres"
echo ""
print_success "Ready to start developing!"


exit 0

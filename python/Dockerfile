# flex-db Python Backend Dockerfile
#
# Multi-stage build for optimized image size and security
#

# Stage 1: Build stage
FROM python:3.11-slim AS builder

WORKDIR /app

# Install build dependencies for psycopg2
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir --user -r requirements.txt

# Stage 2: Production stage
FROM python:3.11-slim

WORKDIR /app

# Install runtime dependencies for PostgreSQL client
RUN apt-get update && apt-get install -y --no-install-recommends \
    libpq5 \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /var/cache/apt/archives/*

# Create a non-root user for security
RUN useradd --create-home --shell /bin/bash appuser

# Copy installed packages from builder stage
COPY --from=builder /root/.local /home/appuser/.local

# Copy application code
COPY --chown=appuser:appuser . .

# Switch to non-root user
USER appuser

# Set HOME and add local packages to PATH
# HOME must be set so Python can find user-site packages in /home/appuser/.local
ENV HOME=/home/appuser \
    PATH=/home/appuser/.local/bin:$PATH

# Set Python environment variables
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

# Default environment variables (can be overridden at runtime)
# Note: DB_PASSWORD should be provided at runtime via -e or docker-compose
ENV DB_HOST=localhost \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_NAME=dbaas \
    DB_SSL_MODE=disable \
    JSONRPC_HOST=0.0.0.0 \
    JSONRPC_PORT=5000

# Expose the JSON-RPC port
EXPOSE 5000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import os, urllib.request; port = os.getenv('JSONRPC_PORT', '5000'); urllib.request.urlopen(f'http://localhost:{port}/health')" || exit 1

# Run the application
CMD ["python", "main.py"]

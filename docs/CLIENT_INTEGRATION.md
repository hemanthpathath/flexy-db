# Client Integration Guide

This guide explains how clients and other services can discover and integrate with the flex-db gRPC service.

## Overview

There are several ways clients can discover the API structure and make calls:

1. **Proto File Sharing** (Recommended for production)
2. **gRPC Reflection** (Great for development/testing)
3. **Generated Client Code** (Best for Go clients)
4. **API Documentation** (Reference material)

## Method 1: Proto File Sharing (Recommended)

### What to Share

Share the **`.proto` file** with clients. This is the source of truth for your API contract.

**File to share:** `api/proto/dbaas.proto`

### How Clients Use It

#### For Go Clients

1. **Get the proto file** (via git submodule, package manager, or direct copy)
2. **Generate client code:**

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate client code
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/dbaas.proto
```

3. **Use the generated client:**

```go
package main

import (
    "context"
    "log"
    
    pb "your-package/api/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Connect to the server
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // Create a client
    client := pb.NewTenantServiceClient(conn)

    // Make a call
    ctx := context.Background()
    resp, err := client.CreateTenant(ctx, &pb.CreateTenantRequest{
        Slug: "acme-corp",
        Name: "Acme Corporation",
    })
    if err != nil {
        log.Fatalf("Failed to create tenant: %v", err)
    }
    
    log.Printf("Created tenant: %+v", resp.Tenant)
}
```

#### For Other Languages

The same `.proto` file can be used to generate clients in:
- **Python**: `protoc --python_out=. --grpc_python_out=. dbaas.proto`
- **JavaScript/TypeScript**: Use `@grpc/proto-loader` or `grpc-tools`
- **Java**: `protoc --java_out=. --grpc-java_out=. dbaas.proto`
- **C#**: `protoc --csharp_out=. --grpc_csharp_out=. dbaas.proto`
- **Ruby**: `protoc --ruby_out=. --grpc_ruby_out=. dbaas.proto`

See [gRPC documentation](https://grpc.io/docs/languages/) for language-specific instructions.

#### For Testing Tools (Insomnia, Postman, BloomRPC)

1. Import the `.proto` file directly
2. The tool parses it and shows all available services and methods
3. You can make requests with autocomplete for request fields

## Method 2: gRPC Reflection (Development/Testing)

### What is gRPC Reflection?

gRPC Reflection allows clients to **dynamically discover** services and methods at runtime without needing the `.proto` file upfront.

**Status:** ✅ **Enabled** on your server (see `cmd/dbaas-server/main.go` line 76)

### How Clients Use It

#### Using grpcurl (Command Line)

```bash
# List all available services
grpcurl -plaintext localhost:50051 list

# Output:
# dbaas.TenantService
# dbaas.UserService
# dbaas.NodeTypeService
# dbaas.NodeService
# dbaas.RelationshipService

# List methods for a service
grpcurl -plaintext localhost:50051 list dbaas.TenantService

# Output:
# dbaas.TenantService.CreateTenant
# dbaas.TenantService.GetTenant
# dbaas.TenantService.UpdateTenant
# dbaas.TenantService.DeleteTenant
# dbaas.TenantService.ListTenants

# Describe a method (see request/response structure)
grpcurl -plaintext localhost:50051 describe dbaas.TenantService.CreateTenant

# Make a call (reflection provides the schema)
grpcurl -plaintext -d '{"slug": "acme", "name": "Acme Corp"}' \
  localhost:50051 dbaas.TenantService/CreateTenant
```

#### Using Insomnia/Postman with Reflection

1. Connect to `localhost:50051`
2. Enable "Use gRPC Reflection"
3. The tools will automatically discover all services and methods
4. You can browse and call methods without importing the proto file

#### Using evans (Interactive REPL)

```bash
evans --host localhost --port 50051 -r repl

# Inside evans:
# > show service
# > call CreateTenant
```

### When to Use Reflection

- ✅ **Development & Testing**: Quick exploration, no proto file needed
- ✅ **Debugging**: Inspect available services dynamically
- ❌ **Production**: Security risk (exposes API structure), performance overhead

**Recommendation:** Keep reflection enabled for development, disable in production.

## Method 3: Generated Client Code (Go Clients)

### Option A: Share Generated Code as a Go Module

If clients are also in Go, you can publish the generated client code as a Go module:

```go
// Client's go.mod
module my-client

require (
    github.com/hemanthpathath/flex-db/api/proto v1.0.0
)
```

Then clients can import and use:
```go
import pb "github.com/hemanthpathath/flex-db/api/proto"

client := pb.NewTenantServiceClient(conn)
```

### Option B: Client Generates from Proto

Clients download the proto file (from git, HTTP, or package manager) and generate their own client code.

## Method 4: API Documentation

### Auto-generated from Proto

You can generate API documentation from the `.proto` file using tools like:

- **protoc-gen-doc**: Generates HTML/Markdown docs
- **grpcurl describe**: Command-line method descriptions
- **buf**: Modern protobuf tooling with documentation generation

### Manual Documentation

Keep API documentation in markdown (like this guide) for reference.

## Distribution Methods for Proto Files

### Option 1: Git Repository

```bash
# Clients can clone or use as git submodule
git clone https://github.com/hemanthpathath/flex-db.git
# Use api/proto/dbaas.proto
```

### Option 2: HTTP/HTTPS Endpoint

Serve the proto file via HTTP:
```bash
# Your server could serve:
GET https://api.yourdomain.com/proto/dbaas.proto
```

### Option 3: Package Manager

- **Go**: Use Go modules to share proto definitions
- **npm**: Publish proto files as npm package
- **Maven/Gradle**: For Java projects
- **PyPI**: For Python projects

### Option 4: Artifact Registry

Use artifact registries like:
- **Buf Schema Registry** (buf.build)
- **GitHub Packages**
- **Docker Hub** (for containerized distribution)

## Example: Complete Go Client

```go
package main

import (
    "context"
    "log"
    "time"
    
    pb "github.com/hemanthpathath/flex-db/api/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Connect to server
    conn, err := grpc.Dial(
        "localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // Create clients for different services
    tenantClient := pb.NewTenantServiceClient(conn)
    userClient := pb.NewUserServiceClient(conn)
    nodeClient := pb.NewNodeServiceClient(conn)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Example: Create a tenant
    tenantResp, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
        Slug: "acme-corp",
        Name: "Acme Corporation",
    })
    if err != nil {
        log.Fatalf("Failed to create tenant: %v", err)
    }
    log.Printf("Created tenant: %s (ID: %s)", tenantResp.Tenant.Name, tenantResp.Tenant.Id)

    // Example: List tenants
    listResp, err := tenantClient.ListTenants(ctx, &pb.ListTenantsRequest{
        Pagination: &pb.Pagination{
            PageSize: 10,
        },
    })
    if err != nil {
        log.Fatalf("Failed to list tenants: %v", err)
    }
    log.Printf("Found %d tenants", len(listResp.Tenants))
}
```

## Example: Python Client

```python
import grpc
import dbaas_pb2
import dbaas_pb2_grpc

def main():
    # Connect to server
    channel = grpc.insecure_channel('localhost:50051')
    stub = dbaas_pb2_grpc.TenantServiceStub(channel)

    # Create a tenant
    request = dbaas_pb2.CreateTenantRequest(
        slug="acme-corp",
        name="Acme Corporation"
    )
    response = stub.CreateTenant(request)
    print(f"Created tenant: {response.tenant.name}")

    channel.close()

if __name__ == '__main__':
    main()
```

## Security Considerations

### Production Recommendations

1. **Disable gRPC Reflection** in production:
   ```go
   // Remove or comment out in production:
   // reflection.Register(grpcServer)
   ```

2. **Use TLS/SSL** for connections:
   ```go
   creds, err := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
   grpcServer := grpc.NewServer(grpc.Creds(creds))
   ```

3. **Share proto files securely**:
   - Via authenticated endpoints
   - Through private artifact registries
   - With version control and access control

## Versioning

### Proto File Versioning

Best practices:
1. **Use semantic versioning** in your proto package name or file version comments
2. **Maintain backward compatibility** when possible
3. **Document breaking changes** clearly
4. **Provide migration guides** for major versions

Example versioning in proto:
```protobuf
syntax = "proto3";

package dbaas.v1;  // Version in package name

// Or use comments:
// Version: 1.2.3
// Last Updated: 2024-01-01
```

## Summary

| Method | Use Case | Pros | Cons |
|--------|----------|------|------|
| **Proto File** | Production clients | Standard, versioned, language-agnostic | Need to distribute file |
| **Reflection** | Development/testing | No file needed, dynamic discovery | Security risk, performance overhead |
| **Generated Code** | Go clients | Type-safe, easy integration | Language-specific |
| **Documentation** | Reference | Human-readable | Needs manual updates |

**Best Practice:** Use **proto file sharing** for production clients, keep **reflection enabled** for development/testing.


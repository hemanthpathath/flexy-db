package grpc

import (
	"context"

	pb "github.com/hemanthpathath/flex-db/go/api/proto"
	grpcerrors "github.com/hemanthpathath/flex-db/go/internal/grpc/errors"
	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/hemanthpathath/flex-db/go/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TenantHandler implements the TenantService gRPC server
type TenantHandler struct {
	pb.UnimplementedTenantServiceServer
	svc *service.TenantService
}

// NewTenantHandler creates a new TenantHandler
func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

// CreateTenant creates a new tenant
func (h *TenantHandler) CreateTenant(ctx context.Context, req *pb.CreateTenantRequest) (*pb.CreateTenantResponse, error) {
	tenant, err := h.svc.Create(ctx, req.Slug, req.Name)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.CreateTenantResponse{
		Tenant: tenantToProto(tenant),
	}, nil
}

// GetTenant retrieves a tenant by ID
func (h *TenantHandler) GetTenant(ctx context.Context, req *pb.GetTenantRequest) (*pb.GetTenantResponse, error) {
	tenant, err := h.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.GetTenantResponse{
		Tenant: tenantToProto(tenant),
	}, nil
}

// UpdateTenant updates an existing tenant
func (h *TenantHandler) UpdateTenant(ctx context.Context, req *pb.UpdateTenantRequest) (*pb.UpdateTenantResponse, error) {
	tenant, err := h.svc.Update(ctx, req.Id, req.Slug, req.Name, req.Status)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.UpdateTenantResponse{
		Tenant: tenantToProto(tenant),
	}, nil
}

// DeleteTenant deletes a tenant
func (h *TenantHandler) DeleteTenant(ctx context.Context, req *pb.DeleteTenantRequest) (*pb.DeleteTenantResponse, error) {
	if err := h.svc.Delete(ctx, req.Id); err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.DeleteTenantResponse{}, nil
}

// ListTenants retrieves tenants with pagination
func (h *TenantHandler) ListTenants(ctx context.Context, req *pb.ListTenantsRequest) (*pb.ListTenantsResponse, error) {
	var pageSize int32 = 10
	var pageToken string

	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
		pageToken = req.Pagination.PageToken
	}

	tenants, result, err := h.svc.List(ctx, pageSize, pageToken)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	pbTenants := make([]*pb.Tenant, len(tenants))
	for i, t := range tenants {
		pbTenants[i] = tenantToProto(t)
	}

	return &pb.ListTenantsResponse{
		Tenants: pbTenants,
		Pagination: &pb.PaginationResponse{
			NextPageToken: result.NextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// tenantToProto converts a repository.Tenant to pb.Tenant
func tenantToProto(t *repository.Tenant) *pb.Tenant {
	return &pb.Tenant{
		Id:        t.ID,
		Slug:      t.Slug,
		Name:      t.Name,
		Status:    t.Status,
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}
}

package grpc

import (
	"context"

	pb "github.com/hemanthpathath/flex-db/go/api/proto"
	grpcerrors "github.com/hemanthpathath/flex-db/go/internal/grpc/errors"
	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/hemanthpathath/flex-db/go/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RelationshipHandler implements the RelationshipService gRPC server
type RelationshipHandler struct {
	pb.UnimplementedRelationshipServiceServer
	svc *service.RelationshipService
}

// NewRelationshipHandler creates a new RelationshipHandler
func NewRelationshipHandler(svc *service.RelationshipService) *RelationshipHandler {
	return &RelationshipHandler{svc: svc}
}

// CreateRelationship creates a new relationship
func (h *RelationshipHandler) CreateRelationship(ctx context.Context, req *pb.CreateRelationshipRequest) (*pb.CreateRelationshipResponse, error) {
	rel, err := h.svc.Create(ctx, req.TenantId, req.SourceNodeId, req.TargetNodeId, req.RelationshipType, req.Data)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.CreateRelationshipResponse{
		Relationship: relationshipToProto(rel),
	}, nil
}

// GetRelationship retrieves a relationship by ID
func (h *RelationshipHandler) GetRelationship(ctx context.Context, req *pb.GetRelationshipRequest) (*pb.GetRelationshipResponse, error) {
	rel, err := h.svc.GetByID(ctx, req.TenantId, req.Id)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.GetRelationshipResponse{
		Relationship: relationshipToProto(rel),
	}, nil
}

// UpdateRelationship updates an existing relationship
func (h *RelationshipHandler) UpdateRelationship(ctx context.Context, req *pb.UpdateRelationshipRequest) (*pb.UpdateRelationshipResponse, error) {
	rel, err := h.svc.Update(ctx, req.TenantId, req.Id, req.RelationshipType, req.Data)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.UpdateRelationshipResponse{
		Relationship: relationshipToProto(rel),
	}, nil
}

// DeleteRelationship deletes a relationship
func (h *RelationshipHandler) DeleteRelationship(ctx context.Context, req *pb.DeleteRelationshipRequest) (*pb.DeleteRelationshipResponse, error) {
	if err := h.svc.Delete(ctx, req.TenantId, req.Id); err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.DeleteRelationshipResponse{}, nil
}

// ListRelationships retrieves relationships with pagination
func (h *RelationshipHandler) ListRelationships(ctx context.Context, req *pb.ListRelationshipsRequest) (*pb.ListRelationshipsResponse, error) {
	var pageSize int32 = 10
	var pageToken string

	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
		pageToken = req.Pagination.PageToken
	}

	rels, result, err := h.svc.List(ctx, req.TenantId, req.SourceNodeId, req.TargetNodeId, req.RelationshipType, pageSize, pageToken)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	pbRels := make([]*pb.Relationship, len(rels))
	for i, r := range rels {
		pbRels[i] = relationshipToProto(r)
	}

	return &pb.ListRelationshipsResponse{
		Relationships: pbRels,
		Pagination: &pb.PaginationResponse{
			NextPageToken: result.NextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// relationshipToProto converts a repository.Relationship to pb.Relationship
func relationshipToProto(r *repository.Relationship) *pb.Relationship {
	return &pb.Relationship{
		Id:               r.ID,
		TenantId:         r.TenantID,
		SourceNodeId:     r.SourceNodeID,
		TargetNodeId:     r.TargetNodeID,
		RelationshipType: r.RelationshipType,
		Data:             r.Data,
		CreatedAt:        timestamppb.New(r.CreatedAt),
		UpdatedAt:        timestamppb.New(r.UpdatedAt),
	}
}

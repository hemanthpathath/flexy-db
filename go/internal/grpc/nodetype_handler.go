package grpc

import (
	"context"

	pb "github.com/hemanthpathath/flex-db/go/api/proto"
	grpcerrors "github.com/hemanthpathath/flex-db/go/internal/grpc/errors"
	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/hemanthpathath/flex-db/go/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NodeTypeHandler implements the NodeTypeService gRPC server
type NodeTypeHandler struct {
	pb.UnimplementedNodeTypeServiceServer
	svc *service.NodeTypeService
}

// NewNodeTypeHandler creates a new NodeTypeHandler
func NewNodeTypeHandler(svc *service.NodeTypeService) *NodeTypeHandler {
	return &NodeTypeHandler{svc: svc}
}

// CreateNodeType creates a new node type
func (h *NodeTypeHandler) CreateNodeType(ctx context.Context, req *pb.CreateNodeTypeRequest) (*pb.CreateNodeTypeResponse, error) {
	nodeType, err := h.svc.Create(ctx, req.TenantId, req.Name, req.Description, req.Schema)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.CreateNodeTypeResponse{
		NodeType: nodeTypeToProto(nodeType),
	}, nil
}

// GetNodeType retrieves a node type by ID
func (h *NodeTypeHandler) GetNodeType(ctx context.Context, req *pb.GetNodeTypeRequest) (*pb.GetNodeTypeResponse, error) {
	nodeType, err := h.svc.GetByID(ctx, req.TenantId, req.Id)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.GetNodeTypeResponse{
		NodeType: nodeTypeToProto(nodeType),
	}, nil
}

// UpdateNodeType updates an existing node type
func (h *NodeTypeHandler) UpdateNodeType(ctx context.Context, req *pb.UpdateNodeTypeRequest) (*pb.UpdateNodeTypeResponse, error) {
	nodeType, err := h.svc.Update(ctx, req.TenantId, req.Id, req.Name, req.Description, req.Schema)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.UpdateNodeTypeResponse{
		NodeType: nodeTypeToProto(nodeType),
	}, nil
}

// DeleteNodeType deletes a node type
func (h *NodeTypeHandler) DeleteNodeType(ctx context.Context, req *pb.DeleteNodeTypeRequest) (*pb.DeleteNodeTypeResponse, error) {
	if err := h.svc.Delete(ctx, req.TenantId, req.Id); err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.DeleteNodeTypeResponse{}, nil
}

// ListNodeTypes retrieves node types with pagination
func (h *NodeTypeHandler) ListNodeTypes(ctx context.Context, req *pb.ListNodeTypesRequest) (*pb.ListNodeTypesResponse, error) {
	var pageSize int32 = 10
	var pageToken string

	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
		pageToken = req.Pagination.PageToken
	}

	nodeTypes, result, err := h.svc.List(ctx, req.TenantId, pageSize, pageToken)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	pbNodeTypes := make([]*pb.NodeType, len(nodeTypes))
	for i, nt := range nodeTypes {
		pbNodeTypes[i] = nodeTypeToProto(nt)
	}

	return &pb.ListNodeTypesResponse{
		NodeTypes: pbNodeTypes,
		Pagination: &pb.PaginationResponse{
			NextPageToken: result.NextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// nodeTypeToProto converts a repository.NodeType to pb.NodeType
func nodeTypeToProto(nt *repository.NodeType) *pb.NodeType {
	return &pb.NodeType{
		Id:          nt.ID,
		TenantId:    nt.TenantID,
		Name:        nt.Name,
		Description: nt.Description,
		Schema:      nt.Schema,
		CreatedAt:   timestamppb.New(nt.CreatedAt),
		UpdatedAt:   timestamppb.New(nt.UpdatedAt),
	}
}

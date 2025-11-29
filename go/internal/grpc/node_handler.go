package grpc

import (
	"context"

	pb "github.com/hemanthpathath/flex-db/go/api/proto"
	grpcerrors "github.com/hemanthpathath/flex-db/go/internal/grpc/errors"
	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/hemanthpathath/flex-db/go/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NodeHandler implements the NodeService gRPC server
type NodeHandler struct {
	pb.UnimplementedNodeServiceServer
	svc *service.NodeService
}

// NewNodeHandler creates a new NodeHandler
func NewNodeHandler(svc *service.NodeService) *NodeHandler {
	return &NodeHandler{svc: svc}
}

// CreateNode creates a new node
func (h *NodeHandler) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeResponse, error) {
	node, err := h.svc.Create(ctx, req.TenantId, req.NodeTypeId, req.Data)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.CreateNodeResponse{
		Node: nodeToProto(node),
	}, nil
}

// GetNode retrieves a node by ID
func (h *NodeHandler) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	node, err := h.svc.GetByID(ctx, req.TenantId, req.Id)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.GetNodeResponse{
		Node: nodeToProto(node),
	}, nil
}

// UpdateNode updates an existing node
func (h *NodeHandler) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeResponse, error) {
	node, err := h.svc.Update(ctx, req.TenantId, req.Id, req.Data)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.UpdateNodeResponse{
		Node: nodeToProto(node),
	}, nil
}

// DeleteNode deletes a node
func (h *NodeHandler) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeResponse, error) {
	if err := h.svc.Delete(ctx, req.TenantId, req.Id); err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.DeleteNodeResponse{}, nil
}

// ListNodes retrieves nodes with pagination
func (h *NodeHandler) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	var pageSize int32 = 10
	var pageToken string

	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
		pageToken = req.Pagination.PageToken
	}

	nodes, result, err := h.svc.List(ctx, req.TenantId, req.NodeTypeId, pageSize, pageToken)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	pbNodes := make([]*pb.Node, len(nodes))
	for i, n := range nodes {
		pbNodes[i] = nodeToProto(n)
	}

	return &pb.ListNodesResponse{
		Nodes: pbNodes,
		Pagination: &pb.PaginationResponse{
			NextPageToken: result.NextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// nodeToProto converts a repository.Node to pb.Node
func nodeToProto(n *repository.Node) *pb.Node {
	return &pb.Node{
		Id:         n.ID,
		TenantId:   n.TenantID,
		NodeTypeId: n.NodeTypeID,
		Data:       n.Data,
		CreatedAt:  timestamppb.New(n.CreatedAt),
		UpdatedAt:  timestamppb.New(n.UpdatedAt),
	}
}

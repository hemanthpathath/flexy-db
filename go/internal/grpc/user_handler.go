package grpc

import (
	"context"

	pb "github.com/hemanthpathath/flex-db/go/api/proto"
	grpcerrors "github.com/hemanthpathath/flex-db/go/internal/grpc/errors"
	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/hemanthpathath/flex-db/go/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserHandler implements the UserService gRPC server
type UserHandler struct {
	pb.UnimplementedUserServiceServer
	svc *service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user, err := h.svc.Create(ctx, req.Email, req.DisplayName)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.CreateUserResponse{
		User: userToProto(user),
	}, nil
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.GetUserResponse{
		User: userToProto(user),
	}, nil
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	user, err := h.svc.Update(ctx, req.Id, req.Email, req.DisplayName)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.UpdateUserResponse{
		User: userToProto(user),
	}, nil
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := h.svc.Delete(ctx, req.Id); err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.DeleteUserResponse{}, nil
}

// ListUsers retrieves users with pagination
func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	var pageSize int32 = 10
	var pageToken string

	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
		pageToken = req.Pagination.PageToken
	}

	users, result, err := h.svc.List(ctx, pageSize, pageToken)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	pbUsers := make([]*pb.User, len(users))
	for i, u := range users {
		pbUsers[i] = userToProto(u)
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
		Pagination: &pb.PaginationResponse{
			NextPageToken: result.NextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// AddUserToTenant adds a user to a tenant
func (h *UserHandler) AddUserToTenant(ctx context.Context, req *pb.AddUserToTenantRequest) (*pb.AddUserToTenantResponse, error) {
	tenantUser, err := h.svc.AddToTenant(ctx, req.TenantId, req.UserId, req.Role)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.AddUserToTenantResponse{
		TenantUser: tenantUserToProto(tenantUser),
	}, nil
}

// RemoveUserFromTenant removes a user from a tenant
func (h *UserHandler) RemoveUserFromTenant(ctx context.Context, req *pb.RemoveUserFromTenantRequest) (*pb.RemoveUserFromTenantResponse, error) {
	if err := h.svc.RemoveFromTenant(ctx, req.TenantId, req.UserId); err != nil {
		return nil, grpcerrors.MapError(err)
	}

	return &pb.RemoveUserFromTenantResponse{}, nil
}

// ListTenantUsers lists users in a tenant
func (h *UserHandler) ListTenantUsers(ctx context.Context, req *pb.ListTenantUsersRequest) (*pb.ListTenantUsersResponse, error) {
	var pageSize int32 = 10
	var pageToken string

	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
		pageToken = req.Pagination.PageToken
	}

	tenantUsers, result, err := h.svc.ListTenantUsers(ctx, req.TenantId, pageSize, pageToken)
	if err != nil {
		return nil, grpcerrors.MapError(err)
	}

	pbTenantUsers := make([]*pb.TenantUser, len(tenantUsers))
	for i, tu := range tenantUsers {
		pbTenantUsers[i] = tenantUserToProto(tu)
	}

	return &pb.ListTenantUsersResponse{
		TenantUsers: pbTenantUsers,
		Pagination: &pb.PaginationResponse{
			NextPageToken: result.NextPageToken,
			TotalCount:    int32(result.TotalCount),
		},
	}, nil
}

// userToProto converts a repository.User to pb.User
func userToProto(u *repository.User) *pb.User {
	return &pb.User{
		Id:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		CreatedAt:   timestamppb.New(u.CreatedAt),
		UpdatedAt:   timestamppb.New(u.UpdatedAt),
	}
}

// tenantUserToProto converts a repository.TenantUser to pb.TenantUser
func tenantUserToProto(tu *repository.TenantUser) *pb.TenantUser {
	return &pb.TenantUser{
		TenantId: tu.TenantID,
		UserId:   tu.UserID,
		Role:     tu.Role,
		Status:   tu.Status,
	}
}

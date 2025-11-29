package errors

import (
	"errors"
	"strings"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapError converts domain errors to gRPC status errors
func MapError(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	// Check for validation errors
	errMsg := err.Error()
	if strings.Contains(errMsg, "required") || strings.Contains(errMsg, "invalid") {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

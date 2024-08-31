package db

import (
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *GetProfileRequest) Validate() error {
	if g.InstanceId == "" {
		return status.Error(codes.InvalidArgument, "instanceId is required")
	}
	if g.Type == "" {
		return status.Error(codes.InvalidArgument, "profileType is required")
	}
	return nil
}

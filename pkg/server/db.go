package server

import (
	"bytes"
	"context"
	"time"

	"github.com/alexandreLamarre/pprof-server/pkg/api/db"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ db.DBServer = (*PprofServer)(nil)

func (p *PprofServer) Get(ctx context.Context, req *db.GetProfileRequest) (*db.GetProfileResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	startTime := lo.ToPtr(lo.FromPtrOr(req.Start, *timestamppb.New(time.Unix(0, 0)))).AsTime()
	endTime := lo.ToPtr(lo.FromPtrOr(req.End, *timestamppb.New(time.Now()))).AsTime()
	ret, err := p.store.Get(ctx, req.InstanceId, req.Type, startTime, endTime)
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer([]byte{})
	if err := ret.Write(b); err != nil { // note: this is compressed by default
		return nil, status.Errorf(codes.Internal, "failed to write profile: %s to buffer", err)
	}
	return &db.GetProfileResponse{
		Data: b.Bytes(),
	}, nil

}

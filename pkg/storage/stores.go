package storage

import (
	"context"
	"time"

	"github.com/google/pprof/profile"
)

type ProfileStore interface {
	// TODO : this should be one profile at a time, and maybe more performant implementations while compact on regular intervals
	Put(
		ctx context.Context,
		instanceId,
		profileType string,
		metadata map[string]string,
		profile []*profile.Profile,
	) error
	Get(ctx context.Context, instanceId, profileType string, start, end time.Time) (*profile.Profile, error)
}

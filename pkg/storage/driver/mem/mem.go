package mem

import (
	"context"
	"time"

	"github.com/alexandreLamarre/pprof-server/pkg/storage"
	"github.com/google/pprof/profile"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func rangeFromProfile(prof *profile.Profile) (start, end time.Time) {
	dur := prof.DurationNanos
	profStart := prof.TimeNanos

	return time.Unix(0, profStart), time.Unix(0, profStart+dur)

}

type profileMemStorage struct {
	// id -> storedProfiles
	buffer map[string]*storedProfiles
}

func NewProfileMemStorage() storage.ProfileStore {
	return &profileMemStorage{
		buffer: map[string]*storedProfiles{},
	}
}

type storedProfiles struct {
	Metadata struct{}
	Labels   map[string]string
	// profile type ( mutex, cpu, etc.. ) -> profile
	Profiles map[string][]*profile.Profile
}

func (m *profileMemStorage) Put(ctx context.Context,
	instanceId, profileType string,
	metadata map[string]string,
	prof []*profile.Profile) error {
	if _, ok := m.buffer[instanceId]; !ok {
		m.buffer[instanceId] = &storedProfiles{
			Labels:   map[string]string{},
			Profiles: map[string][]*profile.Profile{},
		}
	}
	if _, ok := m.buffer[instanceId].Profiles[instanceId]; !ok {
		m.buffer[instanceId].Profiles[profileType] = prof

		// TODO : metadata should be segmented per profile, and metadata changes tracked in a separate index
		// TODO : also we need an index to keep track of what ranges of profiles are compatible.
		m.buffer[instanceId].Labels = metadata
	} else {
		// merge profiles
		m.buffer[instanceId].Profiles[profileType] = append(m.buffer[instanceId].Profiles[instanceId], prof...)
	}
	return nil
}
func (m *profileMemStorage) Get(ctx context.Context, instanceId, profileType string, start, end time.Time) (*profile.Profile, error) {
	profs, ok := m.buffer[instanceId]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "instance not found")
	}
	if _, ok := profs.Profiles[profileType]; !ok {
		return nil, status.Errorf(codes.NotFound, "profile type not found for instanceId")
	}
	retProfiles := []*profile.Profile{}
	for _, profile := range profs.Profiles[profileType] {
		pStart, pEnd := rangeFromProfile(profile)

		if pStart.After(end) {
			continue
		}
		if pEnd.Before(start) {
			continue
		}
		retProfiles = append(retProfiles, profile)
	}

	// TODO : block profiles don't play nice with merge, need to check implementation of `-base` flag to see what they do there
	// TODO : also, for good measure, need to check implementation of `diff_base` flag.
	ret, err := profile.Merge(retProfiles)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to merge profiles: %s, profiles are incompatible", err)
	}
	if valid := ret.CheckValid(); valid != nil {
		return nil, status.Error(codes.FailedPrecondition, "invalid profile after merge")
	}
	return ret, nil
}

package server

import (
	"github.com/alexandreLamarre/pprof-server/pkg/api/db"
	"github.com/alexandreLamarre/pprof-server/pkg/storage"
	"github.com/alexandreLamarre/pprof-server/pkg/storage/driver/mem"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

type PprofServer struct {
	collogspb.UnsafeLogsServiceServer
	db.UnsafeDBServer

	// id -> storedProfiles
	store storage.ProfileStore
}

func NewPprofServer() *PprofServer {
	return &PprofServer{
		store: mem.NewProfileMemStorage(),
	}
}

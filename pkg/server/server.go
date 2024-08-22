package server

import (
	"bytes"
	"context"
	"strings"

	// pprofpb "github.com/alexandreLamarre/pprof-server/pkg/api/pprof"
	"github.com/google/pprof/profile"
	"github.com/sirupsen/logrus"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/protobuf/proto"
)

type PprofServer struct {
	collogspb.UnsafeLogsServiceServer
}

var _ collogspb.LogsServiceServer = (*PprofServer)(nil)

func NewPprofServer() *PprofServer {
	return &PprofServer{}
}

func (p *PprofServer) Export(ctx context.Context, request *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	for _, rscL := range request.GetResourceLogs() {
		for _, scopeL := range rscL.GetScopeLogs() {
			for _, record := range scopeL.GetLogRecords() {

				recordMd := record.GetAttributes()
				scopeMd := scopeL.GetScope().GetAttributes()
				rscMd := rscL.GetResource().GetAttributes()

				allAttributes := append(append(recordMd, scopeMd...), rscMd...)
				res := []string{}
				for _, attr := range allAttributes {
					data, err := proto.Marshal(attr)
					if err != nil {
						logrus.Errorf("Failed to marshal attribute: %v", err)
						continue
					}
					res = append(res, string(data))
				}

				logrus.Infof("Got attributes : %s", strings.Join(res, ","))

				body := record.GetBody().GetBytesValue()
				if len(body) == 0 {
					logrus.Warn("Received empty log record")
					continue
				}

				r := bytes.NewReader(body)
				p, err := profile.Parse(r)
				if err != nil {
					logrus.Errorf("Failed to parse profile: %v", err)
					continue
				}
				logrus.Infof("Received profile: %v", p)
			}
		}
	}

	return &collogspb.ExportLogsServiceResponse{}, nil
}

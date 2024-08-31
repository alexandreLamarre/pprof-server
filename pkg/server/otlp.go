package server

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	// pprofpb "github.com/alexandreLamarre/pprof-server/pkg/api/pprof"
	"github.com/alexandreLamarre/otelbpf/receiver/pprofreceiver"
	"github.com/google/pprof/profile"
	"github.com/sirupsen/logrus"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	otlpcommonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var _ collogspb.LogsServiceServer = (*PprofServer)(nil)

func parseMetadata(attr []*otlpcommonv1.KeyValue) (pprofreceiver.Metadata, map[string]string, error) {
	md := pprofreceiver.Metadata{}
	rawMd := map[string]string{}
	for _, kv := range attr {
		switch kv.Key {
		case "pprof_id":
			md.Id = kv.Value.GetStringValue()
		case "pprof_profile_type":
			md.ProfileType = kv.Value.GetStringValue()
		case "pprof_host":
			md.Host = kv.Value.GetStringValue()
		case "pprof_port":
			md.Port = kv.Value.GetStringValue()
		default:
			rawMd[kv.Key] = kv.Value.GetStringValue()
		}
	}
	if md.Id == "" {
		return md, rawMd, fmt.Errorf("missing id, unable to persist sample profile")
	}
	if md.ProfileType == "" {
		return md, rawMd, fmt.Errorf("missing profile type, unable to persist sample profile")
	}
	return md, rawMd, nil
}

func (p *PprofServer) Export(ctx context.Context, request *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	for _, rscL := range request.GetResourceLogs() {
		for _, scopeL := range rscL.GetScopeLogs() {
			for _, record := range scopeL.GetLogRecords() {

				recordMd := record.GetAttributes()
				scopeMd := scopeL.GetScope().GetAttributes()
				rscMd := rscL.GetResource().GetAttributes()

				allAttributes := append(append(recordMd, scopeMd...), rscMd...)

				// ====== start debug stuff
				res := []string{}
				for _, attr := range allAttributes {
					data, err := protojson.Marshal(attr)
					if err != nil {
						logrus.Errorf("Failed to marshal attribute: %v", err)
						continue
					}
					res = append(res, string(data))
				}

				logrus.Infof("Got attributes : %s", strings.Join(res, ","))
				// ======= end debug stuff

				pMd, md, err := parseMetadata(allAttributes)
				if err != nil {
					logrus.Errorf("Failed to parse metadata: %v", err)
					continue
				}

				body := record.GetBody().GetBytesValue()
				if len(body) == 0 {
					logrus.Warn("Received empty log record")
					continue
				}

				r := bytes.NewReader(body)
				prof, err := profile.Parse(r)
				if err != nil {
					logrus.Errorf("Failed to parse profile: %v", err)
					continue
				}

				if valid := prof.CheckValid(); valid != nil {
					logrus.Errorf("Invalid profile: %v", valid)
					continue
				}

				if err := p.store.Put(ctx, pMd.Id, pMd.ProfileType, md, []*profile.Profile{prof}); err != nil {
					logrus.Errorf("Failed to store profile: %v", err)
					continue
				}
			}
		}
	}

	return &collogspb.ExportLogsServiceResponse{}, nil
}

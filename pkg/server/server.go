package server

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"

	// pprofpb "github.com/alexandreLamarre/pprof-server/pkg/api/pprof"
	"github.com/alexandreLamarre/otelbpf/receiver/pprofreceiver"
	"github.com/google/pprof/profile"
	"github.com/google/pprof/public/ui"
	"github.com/sirupsen/logrus"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	otlpcommonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type storedProfiles struct {
	Metadata struct{}
	Labels   map[string]string
	// profile type ( mutex, cpu, etc.. ) -> profile
	Profiles map[string]*profile.Profile
}

type PprofServer struct {
	collogspb.UnsafeLogsServiceServer

	// id -> storedProfiles
	buffer map[string]*storedProfiles
}

var _ collogspb.LogsServiceServer = (*PprofServer)(nil)

func NewPprofServer() *PprofServer {
	return &PprofServer{
		buffer: map[string]*storedProfiles{},
	}
}

func parseMetadata(attr []*otlpcommonv1.KeyValue) (pprofreceiver.Metadata, error) {
	md := pprofreceiver.Metadata{}
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
		}
	}
	if md.Id == "" {
		return md, fmt.Errorf("missing id, unable to persist sample profile")
	}
	if md.ProfileType == "" {
		return md, fmt.Errorf("missing profile type, unable to persist sample profile")
	}
	return md, nil
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

				md, err := parseMetadata(allAttributes)
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

				if _, ok := p.buffer[md.Id]; !ok {
					p.buffer[md.Id] = &storedProfiles{
						Labels:   map[string]string{},
						Profiles: map[string]*profile.Profile{},
					}
				}

				if _, ok := p.buffer[md.Id].Profiles[md.ProfileType]; !ok {
					p.buffer[md.Id].Profiles[md.ProfileType] = prof
					for _, kv := range allAttributes {
						p.buffer[md.Id].Labels[kv.Key] = kv.Value.GetStringValue()
					}
				} else {
					// merge profiles
					newProf, err := profile.Merge([]*profile.Profile{
						p.buffer[md.Id].Profiles[md.ProfileType],
						prof,
					})
					if err != nil {
						logrus.Errorf("Failed to merge profiles: %v", err)
						continue
					}
					if valid := newProf.CheckValid(); valid != nil {
						logrus.Errorf("Invalid profile after merge: %v", valid)
						continue
					}
					logrus.Info("Successfully merged profiles")
					p.buffer[md.Id].Profiles[md.ProfileType] = newProf
				}
			}
		}
	}

	return &collogspb.ExportLogsServiceResponse{}, nil
}

func (p *PprofServer) HttpServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ui/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.SplitN(r.URL.Path[len("/ui/"):], "/", 3)
		if len(pathParts) < 1 {
			http.Error(w, "id not provided", http.StatusBadRequest)
			return
		}

		id := pathParts[0]
		pType := pathParts[1]
		profiles, ok := p.buffer[id]
		if !ok {
			http.Error(w, "profile not found by ID", http.StatusNotFound)
			return
		}

		profile, ok := profiles.Profiles[pType]
		if !ok {
			logrus.Warnf("Only have profiles of type: %s", strings.Join(slices.Collect(maps.Keys(profiles.Profiles)), ","))
			http.Error(w, "profile not found by type", http.StatusNotFound)
			return
		}

		handler := http.StripPrefix(fmt.Sprintf("/ui/%s/%s", id, pType), createHandlerFromProfile(profile))
		// Serve the request using the new handler
		handler.ServeHTTP(w, r)
	})
	return &http.Server{
		Handler: mux,
	}
}

func createHandlerFromProfile(profile *profile.Profile) http.Handler {
	webUI, err := ui.NewWebUI(profile)
	if err != nil {
		// TODO : don't panic
		panic(err)
	}
	return webUI.Handler()
}

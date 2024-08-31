package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/alexandreLamarre/pprof-server/pkg/api/db"
	"github.com/google/pprof/profile"
	"github.com/google/pprof/public/ui"
	"github.com/sirupsen/logrus"
)

type PprofHttpServer struct {
	dbClient   db.DBClient
	listenAddr string
	mux        *http.ServeMux
	httpServer *http.Server
}

func NewHttpServer(
	listenAddr string,
	dbClient db.DBClient,
) *PprofHttpServer {
	return &PprofHttpServer{
		mux:        http.DefaultServeMux,
		listenAddr: listenAddr,
		dbClient:   dbClient,
	}
}

func (p *PprofHttpServer) ListenAndServe() error {
	listener, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler: p.mux,
	}
	p.registerHandlers()

	return server.Serve(listener)
}

func (p *PprofHttpServer) Shutdown(ctx context.Context) error {
	return p.httpServer.Shutdown(ctx)
}

func (p *PprofHttpServer) registerHandlers() {
	p.mux.HandleFunc("/ui/", p.displayProfile)
}

func (p *PprofHttpServer) displayProfile(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.SplitN(r.URL.Path[len("/ui/"):], "/", 3)
	if len(pathParts) < 1 {
		http.Error(w, "id not provided", http.StatusBadRequest)
		return
	}

	id := pathParts[0]
	pType := pathParts[1]

	data, err := p.dbClient.Get(r.Context(), &db.GetProfileRequest{
		InstanceId: id,
		Type:       pType,
	})
	if err != nil {
		logrus.WithError(err).Error("failed to get profile")
		http.Error(w, "failed to get profile", http.StatusInternalServerError)
		return
	}

	prof, err := profile.ParseData(data.Data)
	if err != nil {
		logrus.WithError(err).Error("failed to parse profile")
		http.Error(w, "failed to parse profile", http.StatusInternalServerError)
		return
	}

	handler := http.StripPrefix(fmt.Sprintf("/ui/%s/%s", id, pType), createHandlerFromProfile(prof))
	// Serve the request using the new handler
	handler.ServeHTTP(w, r)
}

// serves on /ui/<id>/<profile_type>
func (p *PprofHttpServer) HttpServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ui/", p.displayProfile)
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

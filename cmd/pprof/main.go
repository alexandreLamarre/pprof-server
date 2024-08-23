package main

import (
	"net"
	"net/http"
	"time"

	"github.com/alexandreLamarre/pprof-server/pkg/server"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func BuildPprofServer() *cobra.Command {
	var grpcAddr string
	var httpAddr string
	cmd := &cobra.Command{
		Use: "pprofserver",
		RunE: func(cmd *cobra.Command, args []string) error {

			gListener, err := net.Listen("tcp4", grpcAddr)
			if err != nil {
				return err
			}

			hListener, err := net.Listen("tcp", httpAddr)
			if err != nil {
				return err
			}

			grpcServer := grpc.NewServer(
				grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
					MinTime:             15 * time.Second,
					PermitWithoutStream: true,
				}),
				grpc.KeepaliveParams(keepalive.ServerParameters{
					Time:    15 * time.Second,
					Timeout: 5 * time.Second,
				}),
				grpc.StatsHandler(otelgrpc.NewServerHandler()),
			)

			pprofServer := server.NewPprofServer()

			grpcServer.RegisterService(&collogspb.LogsService_ServiceDesc, pprofServer)

			errGC := lo.Async(func() error {
				logrus.Infof("Pprof gRPC server listening on %s....", grpcAddr)
				return grpcServer.Serve(gListener)
			})
			var httpServer *http.Server
			errHC := lo.Async(func() error {
				logrus.Infof("Http UI server listening on %s...", httpAddr)
				s := pprofServer.HttpServer()
				httpServer = s
				return s.Serve(hListener)
			})

			select {
			case <-cmd.Context().Done():
				httpServer.Shutdown(cmd.Context())
				grpcServer.GracefulStop()
				return nil
			case err := <-errHC:
				logrus.Errorf("HTTP server error: %v", err)
				grpcServer.GracefulStop()
				return err
			case err := <-errGC:
				httpServer.Shutdown(cmd.Context())
				logrus.Errorf("GRPC server error: %v", err)
				return err
			}
		},
	}
	cmd.Flags().StringVarP(&httpAddr, "http-addr", "a", ":10000", "The address to listen on for HTTP requests.")
	cmd.Flags().StringVarP(&grpcAddr, "grpc-addr", "g", ":10001", "The address to listen on for GRPC requests.")
	return cmd
}

func main() {
	cmd := BuildPprofServer()
	cmd.Execute()
}

package main

import (
	"net"
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
	var listenAddr string
	cmd := &cobra.Command{
		Use: "pprofserver",
		RunE: func(cmd *cobra.Command, args []string) error {

			listener, err := net.Listen("tcp4", listenAddr)
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

			errC := lo.Async(func() error {
				logrus.Infof("Pprof server listening on %s....", listenAddr)
				return grpcServer.Serve(listener)
			})

			select {
			case <-cmd.Context().Done():
				grpcServer.GracefulStop()
				return nil
			case err := <-errC:
				return err
			}
		},
	}

	cmd.Flags().StringVar(&listenAddr, "listen-addr", ":10001", "The address to listen on for GRPC requests.")
	return cmd
}

func main() {
	cmd := BuildPprofServer()
	cmd.Execute()
}

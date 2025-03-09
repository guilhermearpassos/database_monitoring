package main

import (
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/services/ui/ports"
	"github.com/spf13/cobra"
)

var (
	grpcSrvr     string
	frontendAddr string
	UiCmd        = &cobra.Command{
		Use:     "ui",
		Short:   "run dbm ui",
		Long:    "run dbm ui",
		Aliases: []string{},
		Example: "dbm ui",
		RunE:    StartUI,
	}
)

func init() {
	UiCmd.Flags().StringVar(&grpcSrvr, "grpc-addr", "", "")
	UiCmd.Flags().StringVar(&frontendAddr, "frontend-addr", "", "")
}

func StartUI(cmd *cobra.Command, args []string) error {

	cc, err := telemetry.OpenInstrumentedClientConn(grpcSrvr, int(100000))
	if err != nil {
		return err
	}

	server, err := ports.NewServer(cc)
	if err != nil {
		return err
	}
	err = server.StartServer(frontendAddr)
	if err != nil {
		return err
	}
	return nil
}

package main

import (
	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/guilhermearpassos/database-monitoring/internal/services/ui/ports"
	"github.com/spf13/cobra"
)

var (
	grpcSrvr     string
	frontendAddr string
	otlpAddr     string
	tlsEnabled   bool
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
	UiCmd.Flags().StringVar(&otlpAddr, "otlp-addr", "", "")
	UiCmd.Flags().BoolVar(&tlsEnabled, "tls-enabled", false, "")
}

func StartUI(cmd *cobra.Command, args []string) error {
	enabled := true
	if otlpAddr == "" {
		enabled = false
	}
	telemetryConfig := telemetry.TelemetryConfig{
		Enabled: enabled,
		OTLP: telemetry.OTLPConfig{
			Endpoint: otlpAddr,
		},
	}
	err := telemetry.InitTelemetryFromConfig(telemetryConfig)
	if err != nil {
		return err
	}
	cc, err := telemetry.OpenInstrumentedClientConn(grpcSrvr, int(1000000), tlsEnabled)
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

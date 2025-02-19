package main

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/ui/ports"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	cc, err := grpc.NewClient(grpcSrvr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

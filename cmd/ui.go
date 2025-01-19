package main

import (
	"github.com/guilhermearpassos/database-monitoring/internal/services/ui/ports"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	UiCmd = &cobra.Command{
		Use:     "ui",
		Short:   "run dbm ui",
		Long:    "run dbm ui",
		Aliases: []string{},
		Example: "dbm ui",
		RunE:    StartUI,
	}
)

func StartUI(cmd *cobra.Command, args []string) error {
	cc, err := grpc.NewClient("localhost:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	server, err := ports.NewServer(cc)
	if err != nil {
		return err
	}
	err = server.StartServer("localhost:8080")
	if err != nil {
		return err
	}
	return nil
}

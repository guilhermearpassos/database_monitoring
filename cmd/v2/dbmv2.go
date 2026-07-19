package main

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/guilhermearpassos/database-monitoring/internal/bootstrap"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"time"
)

var (
	DbmV2 = &cobra.Command{
		Use:     "all",
		Short:   "run sqlsights as monolith",
		Long:    "run sqlsights as monolith",
		Aliases: []string{},
		Example: "dbm all",
		RunE:    DBMV2,
	}
	pgAddr string
)

func init() {
}

func DBMV2(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	var cfg bootstrap.AppInstanceConfig
	// Check if file exists
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		panic(fmt.Errorf("config file does not exist: %s", configFileName))
	}
	if _, err := toml.DecodeFile(configFileName, &cfg); err != nil {
		panic(fmt.Errorf("failed to parse config file: %s", err))
	}
	app := bootstrap.NewApplicationInstance(cfg)
	if err := app.Start(ctx); err != nil {
		panic(err)
	}
	<-ctx.Done()
	gracefulCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := app.Stop(gracefulCtx)
	if err != nil {
		panic(err)
	}
	return nil
}

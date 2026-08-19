package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/guilhermearpassos/database-monitoring/internal/bootstrap"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
)

func init() {
	DbmV2.Flags().StringVar(&configFileName, "config", "local/v2.yaml", "--config=local/v2.yaml")
}

func DBMV2(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	var cfg bootstrap.AppInstanceConfig
	// Check if file exists
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		panic(fmt.Errorf("config file does not exist: %s", configFileName))
	}
	file, err := os.Open(configFileName)
	if err != nil {
		panic(fmt.Errorf("failed to open config file: %s", err))
	}
	defer file.Close()
	switch {
	case strings.HasSuffix(configFileName, ".toml"):
		if _, err := toml.NewDecoder(file).Decode(&cfg); err != nil {
			panic(fmt.Errorf("failed to parse TOML config file: %s", err))
		}
	case strings.HasSuffix(configFileName, ".yaml") || strings.HasSuffix(configFileName, ".yml"):
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&cfg); err != nil {
			panic(fmt.Errorf("failed to parse YAML config file: %s", err))
		}
	default:
		panic(fmt.Errorf("unsupported config file format: %s", configFileName))
	}
	app := bootstrap.NewApplicationInstance(ctx, cfg)
	if err := app.Start(ctx); err != nil {
		panic(err)
	}
	<-ctx.Done()
	gracefulCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = app.Stop(gracefulCtx)
	if err != nil {
		panic(err)
	}
	return nil
}

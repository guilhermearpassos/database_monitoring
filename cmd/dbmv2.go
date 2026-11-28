package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/guilhermearpassos/database-monitoring/internal/bootstrap"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	dbmV2ConfigFile string
	DbmV2           = &cobra.Command{
		Use:     "all",
		Short:   "run sqlsights as monolith",
		Long:    "run sqlsights as monolith",
		Aliases: []string{},
		Example: "dbm all",
		RunE:    DBMV2,
	}
)

func init() {
	DbmV2.Flags().StringVar(&dbmV2ConfigFile, "config", "local/v2.yaml", "--config=local/v2.yaml")
}

func DBMV2(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var cfg bootstrap.AppInstanceConfig
	// Check if file exists
	if _, err := os.Stat(dbmV2ConfigFile); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", dbmV2ConfigFile)
	}
	file, err := os.Open(dbmV2ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to open config file: %s: %w", dbmV2ConfigFile, err)
	}
	defer file.Close()
	switch {
	case strings.HasSuffix(dbmV2ConfigFile, ".toml"):
		if _, err := toml.NewDecoder(file).Decode(&cfg); err != nil {
			return fmt.Errorf("failed to parse TOML config file: %s: %w", dbmV2ConfigFile, err)
		}
	case strings.HasSuffix(dbmV2ConfigFile, ".yaml") || strings.HasSuffix(dbmV2ConfigFile, ".yml"):
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&cfg); err != nil {
			return fmt.Errorf("failed to parse YAML config file: %s: %w", dbmV2ConfigFile, err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s", dbmV2ConfigFile)
	}
	app := bootstrap.NewApplicationInstance(ctx, cfg)
	if err := app.Start(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	fmt.Println(ctx.Err())
	gracefulCtx, gracefulCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer gracefulCancel()
	err = app.Stop(gracefulCtx)
	if err != nil {
		return err
	}
	return nil
}

package main

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/guilhermearpassos/database-monitoring/sql"
	"github.com/spf13/cobra"
)

var (
	MigrateCmd = &cobra.Command{
		Use:     "migrate",
		Short:   "run database migrations",
		Long:    "run database migrations",
		Aliases: []string{},
		Example: "dbm migrate",
		RunE:    erMigrate,
	}
	pgAddr string
)

func init() {

	MigrateCmd.Flags().StringVar(&pgAddr, "pg-addr", "", "")
}

func erMigrate(cmd *cobra.Command, args []string) error {
	fmt.Println("migrate")
	d, err := iofs.New(sql.MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrate fs: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, pgAddr)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

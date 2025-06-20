package main

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	m, err := migrate.New("file://./sql/migrations", pgAddr)
	if err != nil {
		panic(err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
	return nil
}

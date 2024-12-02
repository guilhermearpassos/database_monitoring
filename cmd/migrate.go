package main

//import (
//	"github.com/golang-migrate/migrate/v4"
//	_ "github.com/golang-migrate/migrate/v4/database/sqlserver"
//	_ "github.com/golang-migrate/migrate/v4/source/file"
//	"github.com/spf13/cobra"
//)
//
//var (
//	MigrateCmd = &cobra.Command{
//		Use:     "migrate",
//		Short:   "run database migrations",
//		Long:    "run database migrations",
//		Aliases: []string{},
//		Example: "er migrate",
//		RunE:    erMigrate,
//	}
//)
//
//func erMigrate(cmd *cobra.Command, args []string) error {
//	m, err := migrate.New("file://../sql/migrations", "sqlserver://sa:SqlServer2019!@localhost:1433?database=SQL_EXECUTION_ROUTER")
//	if err != nil {
//		panic(err)
//	}
//	err = m.Up()
//	if err != nil && err != migrate.ErrNoChange {
//		panic(err)
//	}
//	return nil
//}

//go:debug x509negativeserial=1
package main

import (
	"github.com/spf13/cobra"
)

func main() {
	c := &cobra.Command{
		Use: "dbm",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	c.AddCommand(DbmV2)
	c.AddCommand(MigrateCmd)
	c.AddCommand(LoadGenCmd)
	err := c.Execute()
	if err != nil {
		panic(err)
	}
}

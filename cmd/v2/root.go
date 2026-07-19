package main

import (
	"github.com/spf13/cobra"
)

var configFileName string

func main() {
	c := &cobra.Command{
		Use: "dbmv2",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	c.AddCommand(DbmV2)
	err := c.Execute()
	if err != nil {
		panic(err)
	}
}

//go:debug x509negativeserial=1
package main

import (
	"github.com/spf13/cobra"
)

var configFileName string

func main() {
	c := &cobra.Command{
		Use: "dbm",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	c.AddCommand(CollectorCmd)
	c.AddCommand(AgentCmd)
	c.AddCommand(GrpcCmd)
	c.AddCommand(UiCmd)
	err := c.Execute()
	if err != nil {
		panic(err)
	}
}

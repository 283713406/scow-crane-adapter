package main

import (
	"github.com/spf13/cobra"
	"scow-crane-adapter/pkg/adapter"
)

func main() {
	rootCmd := adapter.NewAdapterCommand()
	if err := rootCmd.Execute(); err != nil {
		cobra.CheckErr(err)
	}
}

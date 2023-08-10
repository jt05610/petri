/*
Copyright © 2023 Jonathan Taylor <jonrtaylor12@gmail.com>
*/

package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the root command
var rootCmd = &cobra.Command{
	Use:   "modbus",
	Short: "modbus allows interaction with modbus devices",
	Long:  ``,
}

func init() {
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

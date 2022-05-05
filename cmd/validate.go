/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the Keel schema",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("validate called")

		// upgrade the command meta data above

		// harvest input schema file name from cmd arguments

		// slurp the file contents into a string

		// delegate to exported Schema method

		// output something useful
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

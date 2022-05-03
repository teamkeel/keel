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

	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/schema"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your Keel schema",
	Run: func(cmd *cobra.Command, args []string) {

		// This function call not only validates the schemas in your input directory,
		// but also returns the
		// protobuf representation of it. However in this Validate use-case - we
		// take no interest in the returned protobuf models.
		_, err := schema.NewSchema(inputDir).Make()

		if err != nil {
			fmt.Printf("Validation error: %v\n", err)
			return
		}

		fmt.Printf("Validation OK\n")
	},
}

var inputDir string

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVar(&inputDir, "d", "input-dir", "input directory to validate")
}

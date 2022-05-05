/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/schema"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your Keel schema",
	Run: func(cmd *cobra.Command, args []string) {

		schemaBytes, err := ioutil.ReadFile(schemaFilename)
		if err != nil {
			fmt.Printf("Error reading input schema file: <%s> : %v\n", schemaFilename, err)
			return
		}

		_, err = schema.NewSchema(string(schemaBytes)).Make()

		if err != nil {
			fmt.Printf("Validation error: %v\n", err)
			return
		}

		fmt.Printf("Validation OK\n")
	},
}

var schemaFilename string

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVar(&schemaFilename, "file", "keel.schema", "input file to validate")
}

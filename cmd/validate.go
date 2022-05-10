/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/validation"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your Keel schema",
	RunE: func(cmd *cobra.Command, args []string) error {

		schema := schema.Schema{}
		var protoSchema *proto.Schema // For clarity only.
		_ = protoSchema
		var err error
		var errors []error

		switch {
		case inputFile != "":
			protoSchema, errors, err = schema.MakeFromFile(inputFile)
		default:
			protoSchema, errors, err = schema.MakeFromDirectory(inputDir)
		}

		if err != nil {
			_, ok := err.(*validation.ValidationError)
			if ok {
				if output == "json" {
					output, err := json.Marshal(errors)
					if err != nil {
						return fmt.Errorf("error marshalling validation errors: %v", err)
					}
					fmt.Println(string(output))
					return nil
				} else {
					for _, e := range errors {
						fmt.Println(e.Error())
					}
					return nil
				}
			} else {
				return fmt.Errorf("error making schema: %v", err)
			}
		}
		fmt.Printf("Validation OK\n")

		return nil
	},
}

var inputDir string
var inputFile string
var output string

func init() {
	rootCmd.AddCommand(validateCmd)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	validateCmd.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "input directory to validate")
	validateCmd.Flags().StringVarP(&inputFile, "file", "f", "", "schema file to validate")
	validateCmd.Flags().StringVarP(&output, "output", "o", "console", "output format (console, json)")
}

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/pkg/output"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/validation"
)

type validateCommand struct {
	outputFormatter *output.Output
}

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your Keel schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := &validateCommand{
			outputFormatter: output.New(os.Stdout),
		}

		formatter(outputFormat, *c)

		schema := schema.Schema{}
		var protoSchema *proto.Schema // For clarity only.
		_ = protoSchema
		var err error

		switch {
		case inputFile != "":
			protoSchema, err = schema.MakeFromFile(inputFile)
		default:
			protoSchema, err = schema.MakeFromDirectory(inputDir)
		}

		if err != nil {
			errs, ok := err.(validation.ValidationErrors)
			if ok {
				return c.outputFormatter.Write(errs.Errors)
			} else {
				return fmt.Errorf("error making schema: %v", err)
			}
		}
		c.outputFormatter.Write("Validation OK\n")

		return nil
	},
}

func formatter(outputFormatter string, c validateCommand) {
	switch outputFormatter {
	case string(output.FormatterJSON):
		c.outputFormatter.SetOutput(output.FormatterJSON, os.Stdout)
	default:
		c.outputFormatter.SetOutput(output.FormatterConsole, os.Stdout)
	}

}

var inputDir string
var inputFile string
var outputFormat string

func init() {
	rootCmd.AddCommand(validateCmd)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	validateCmd.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "input directory to validate")
	validateCmd.Flags().StringVarP(&inputFile, "file", "f", "", "schema file to validate")
	validateCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "output format (console, json)")
}

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/proto"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your Keel schema",
	Run: func(cmd *cobra.Command, args []string) {

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
			fmt.Printf("Validation error: %v\n", err)
			return
		}

		fmt.Printf("Validation OK\n")
	},
}

var inputDir string
var inputFile string

func init() {
	rootCmd.AddCommand(validateCmd)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	validateCmd.Flags().StringVar(&inputDir, "d", defaultDir, "input directory to validate")
	validateCmd.Flags().StringVar(&inputFile, "f", "", "schema file to validate")
}

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/internal/proto"
	"github.com/teamkeel/keel/parser"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the Keel schema",
	Run: func(cmd *cobra.Command, args []string) {
		schema := `
	model cookTook {
		fields {
		  title Text
		  isbn Text {
			@unique
		  }
		  authors Author[]
		}
		functions {
		  create createBook(title, authors)
		  get book(id)
		  get bookByIsbn(isbn)
		}
	  }`

		res, err := parser.Parse(schema)
		if err != nil {
			panic(err)
		}

		p, err := proto.ToProto(res)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(p)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Read DB migrations directory, construct the schema and diff the two",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("diff called")
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

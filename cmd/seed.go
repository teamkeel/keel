package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var (
	dryRun bool
)

func init() {
	rootCmd.AddCommand(seedCmd)
	seedCmd.Args = cobra.MaximumNArgs(1)
	seedCmd.ValidArgs = []string{"snapshot"}

	seedCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the changes that would be made without applying them")
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Apply seed data to the database",
	Long:  `Executes all SQL files in the seed directory`,
	Run: func(cmd *cobra.Command, args []string) {
		packageManager, err := resolvePackageManager(flagProjectDir, false)
		if err == promptui.ErrAbort {
			return
		}
		if err != nil {
			panic(err)
		}

		shouldSnapshot := false

		if len(args) > 0 {
			if args[0] == "snapshot" {
				shouldSnapshot = true
			}
		}

		program.Run(&program.Model{
			Mode:             program.ModeSeed,
			ProjectDir:       flagProjectDir,
			PackageManager:   packageManager,
			SeedData:         true,
			SnapshotDatabase: shouldSnapshot,
		})
	},
}

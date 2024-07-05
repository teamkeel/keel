package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/runtime"
)

var enabledDebugFlags = "true"

var (
	flagProjectDir       string
	flagReset            bool
	flagPort             string
	flagNodePackagesPath string
	flagPrivateKeyPath   string
	flagPattern          string
	flagTracing          bool
	flagVersion          bool
	flagVerboseTracing   bool
	flagEnvironment      string
	flagHostname         string
	flagJsonOutput       bool
)

var rootCmd = &cobra.Command{
	Use:   "keel",
	Short: "The Keel CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagVersion {
			fmt.Printf("v%s\n", runtime.GetVersion())
			os.Exit(0)
		}
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringVarP(&flagProjectDir, "dir", "d", workingDir, "directory containing a Keel project")
	rootCmd.PersistentFlags().BoolVarP(&flagVersion, "version", "v", false, "print the Keel CLI version")
}

func resolvePackageManager(dir string, isInit bool) (string, error) {
	dir = filepath.Clean(dir)

	for {
		packageLockPath := filepath.Join(dir, "package-lock.json")
		pnpmLockPath := filepath.Join(dir, "pnpm-lock.yaml")

		_, err := os.Stat(packageLockPath)
		if err == nil {
			if isInit {
				fmt.Println("|", colors.Gray("package-lock.json found at"), packageLockPath)
			}
			return "npm", nil
		}

		_, err = os.Stat(pnpmLockPath)
		if err == nil {
			if isInit {
				fmt.Println("|", colors.Gray("pnpm-lock.yaml found at"), packageLockPath)
			}
			return "pnpm", nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	_, err := exec.LookPath("pnpm")
	if err != nil {
		if isInit {
			fmt.Println("|", colors.Gray("pnpm not detected"))
		}
		return "npm", nil
	}

	if !isInit {
		fmt.Println(
			colors.Yellow("| We're not sure which package manager you'd like to use as we can't find any lockfiles"),
		)
		fmt.Println("")
	}

	s := promptui.Select{
		Label: "Which Node.js package manager would you like to use?",
		Items: []string{"pnpm", "npm"},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Selected: fmt.Sprintf("%s {{ . | bold }} ", promptui.IconGood),
			Active:   "{{ . | cyan }}",
			Inactive: "{{ . }}",
		},
	}

	_, v, err := s.Run()
	return v, err
}

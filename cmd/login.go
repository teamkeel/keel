package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/config"
	"github.com/teamkeel/keel/cmd/web"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login to the keel backend",
	Long:  `login to connect to your keel projects from the command line`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := config.New()
		c := &web.Controller{
			Cfg: config,
		}

		_, err := c.Login(context.Background())
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

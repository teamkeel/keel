package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unexpected arguments: %v", args)
		}

		b := schema.Builder{}
		_, err := b.MakeFromDirectory(flagProjectDir)
		if err == nil {
			fmt.Println("✨ Everything's looking good!")
			return nil
		}

		validationErrors := &errorhandling.ValidationErrors{}
		configErrors := &config.ConfigErrors{}

		switch {
		case errors.As(err, &validationErrors):
			fmt.Println("❌ The following errors were found in your schema files:")
			fmt.Println("")
			s := validationErrors.ToAnnotatedSchema(b.SchemaFiles())
			fmt.Println(s)
			return nil
		case errors.As(err, &configErrors):
			fmt.Println("❌ The following errors were found in your", colors.Yellow("keelconfig.yaml").String(), "file:")
			fmt.Println("")
			for _, v := range configErrors.Errors {
				fmt.Println(" -", colors.Red(v.Message).String())
			}
		default:
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

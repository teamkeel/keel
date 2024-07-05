package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

type JsonResponse struct {
	ValidationErrors errorhandling.ValidationErrors `json:"validationErrors"`
	ConfigErrors     config.ConfigErrors            `json:"configErrors"`
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		b := schema.Builder{}
		_, err := b.MakeFromDirectory(flagProjectDir)
		if err == nil {
			fmt.Println("✨ Everything's looking good!")
			return nil
		}

		validationErrors := &errorhandling.ValidationErrors{}
		configErrors := &config.ConfigErrors{}

		if flagJsonOutput {
			var resp JsonResponse
			switch {
			case errors.As(err, &validationErrors):
				resp = JsonResponse{
					ValidationErrors: *validationErrors,
				}
			case errors.As(err, &configErrors):
				resp = JsonResponse{
					ConfigErrors: *configErrors,
				}
			default:
				return err
			}

			json, err := json.Marshal(resp)
			if err != nil {
				return err
			}

			fmt.Println(string(json))

			return nil
		}

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
	validateCmd.Flags().BoolVar(&flagJsonOutput, "json", false, "output validation and config errors as json")
}

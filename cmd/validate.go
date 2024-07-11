package cmd

import (
	"encoding/base64"
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

		var err error
		if flagSchema != "" || flagConfig != "" {
			var schema []byte
			schema, err = base64.StdEncoding.DecodeString(flagSchema)
			if err != nil {
				return err
			}

			var config []byte
			config, err = base64.StdEncoding.DecodeString(flagConfig)
			if err != nil {
				return err
			}

			_, err = b.MakeFromString(string(schema), string(config))
		} else {
			_, err = b.MakeFromDirectory(flagProjectDir)
		}

		if err == nil && !flagJsonOutput {
			fmt.Println("✨ Everything's looking good!")
			return nil
		}

		validationErrors := &errorhandling.ValidationErrors{
			Errors:   []*errorhandling.ValidationError{},
			Warnings: []*errorhandling.ValidationError{},
		}

		configErrors := &config.ConfigErrors{
			Errors: []*config.ConfigError{},
		}

		if flagJsonOutput {
			resp := JsonResponse{
				ValidationErrors: *validationErrors,
				ConfigErrors:     *configErrors,
			}

			switch {
			case errors.As(err, &validationErrors):
				resp.ValidationErrors = *validationErrors
			case errors.As(err, &configErrors):
				resp.ConfigErrors = *configErrors
			default:
				if err != nil {
					return err
				}
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
	validateCmd.Flags().StringVar(&flagSchema, "schema", "", "the Keel schema as base64 passed as an argument")
	validateCmd.Flags().StringVar(&flagConfig, "config", "", "the Keel config as base64 passed as an argument")
}

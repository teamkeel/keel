package cmd

import (
	"encoding/base64"
	"encoding/json"
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

		var validationErrors *errorhandling.ValidationErrors
		var configFiles []*config.ConfigFile

		if flagSchema != "" || flagConfig != "" {
			schema, err := base64.StdEncoding.DecodeString(flagSchema)
			if err != nil {
				return err
			}

			configBytes, err := base64.StdEncoding.DecodeString(flagConfig)
			if err != nil {
				return err
			}

			_, err = b.MakeFromString(string(schema), string(configBytes))
			if err != nil {
				if _, ok := err.(*errorhandling.ValidationErrors); !ok {
					return err
				}

				validationErrors = err.(*errorhandling.ValidationErrors)
			}

			c, err := config.LoadFromBytes(configBytes, "")
			if err != nil {
				if config.ToConfigErrors(err) == nil {
					return err
				}

				configFiles = []*config.ConfigFile{
					{
						// TODO: ideally the VSCode extension would send all config files but for now we'll assume it's just the default one
						Filename: "keelconfig.yaml",
						Env:      "",
						Config:   c,
						Errors:   config.ToConfigErrors(err),
					},
				}
			}

		} else {
			_, err := b.MakeFromDirectory(flagProjectDir)
			if err != nil {
				if _, ok := err.(*errorhandling.ValidationErrors); !ok {
					return err
				}

				validationErrors = err.(*errorhandling.ValidationErrors)
			}

			configFiles, err = config.LoadAll(flagProjectDir)
			if err != nil {
				return nil
			}
		}

		if flagJsonOutput {
			resp := JsonResponse{}
			if validationErrors != nil {
				resp.ValidationErrors = *validationErrors
			}
			for _, f := range configFiles {
				if f.Errors != nil {
					resp.ConfigErrors.Errors = append(resp.ConfigErrors.Errors, f.Errors.Errors...)
				}
			}

			json, err := json.Marshal(resp)
			if err != nil {
				return err
			}

			fmt.Println(string(json))
			return nil
		}

		hasConfigErrors := false
		for _, f := range configFiles {
			if f.Errors != nil && len(f.Errors.Errors) > 0 {
				hasConfigErrors = true
			}
		}

		if validationErrors == nil && !hasConfigErrors {
			fmt.Println("✨ Everything's looking good!")
			return nil
		}

		if validationErrors != nil {
			fmt.Println("❌ The following errors were found in your schema files:")
			fmt.Println("")
			s := validationErrors.ErrorsToAnnotatedSchema(b.SchemaFiles())
			fmt.Println(s)
		}

		for _, f := range configFiles {
			if f.Errors == nil || len(f.Errors.Errors) == 0 {
				continue
			}

			fmt.Printf("❌ The following errors were found in %s:\n", colors.Heading(f.Filename).String())
			fmt.Println("")
			for j, v := range f.Errors.Errors {
				if j > 0 {
					fmt.Println("")
				}
				fmt.Println(" -", colors.Yellow(v.Message).String())
				if v.AnnotatedSource != "" {
					fmt.Println(v.AnnotatedSource)
				}
			}
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

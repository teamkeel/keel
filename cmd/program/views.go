package program

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func renderValidate(m *Model) string {
	b := strings.Builder{}

	if m.Err == nil && m.Schema == nil {
		b.WriteString("⏳ Loading schema")
	}
	if m.Err == nil && m.Schema != nil {
		b.WriteString("✨ Everything's looking good!")
	}

	return b.String()
}

func renderTest(m *Model) string {
	b := strings.Builder{}

	if m.TestOutput != "" {
		b.WriteString(m.TestOutput)
	} else {
		switch m.Status {
		case StatusRunning:
			b.WriteString("🏃‍♂️ Running tests\n")
		default:
			b.WriteString("⏳ Setting up tests\n")
		}
	}

	return b.String()
}

func renderRun(m *Model) string {
	b := strings.Builder{}
	if m.Status == StatusQuitting {
		b.WriteString("Goodbye 👋")
		return b.String() + "\n"
	}

	b.WriteString("\n")
	b.WriteString("Running Keel app in directory: ")
	b.WriteString(colors.White(m.ProjectDir).String())
	b.WriteString("\n")

	if m.DatabaseConnInfo != nil {
		b.WriteString("Connect to your database using: ")
		b.WriteString(colors.Cyan(fmt.Sprintf("psql -Atx \"%s\"", m.DatabaseConnInfo.String())).Highlight().String())
		b.WriteString("\n")
	}

	b.WriteString("\n")

	switch m.Status {
	case StatusSetupDatabase:
		b.WriteString("⏳ Setting up database\n")

	case StatusSetupFunctions:
		b.WriteString("✅ Database running\n")
		b.WriteString("⏳ Checking function setup\n")

	case StatusLoadSchema:
		if m.Err == nil {
			b.WriteString("⏳ Loading schema\n")
		} else {
			b.WriteString("❌ Schema\n")
		}

	case StatusRunMigrations:
		b.WriteString("✅ Schema\n")
		if m.Err == nil {
			b.WriteString("⏳ Database migrations\n")
		} else {
			b.WriteString("❌ Database migrations\n")
		}

	case StatusUpdateFunctions, StatusStartingFunctions:
		b.WriteString("✅ Schema\n")
		b.WriteString("✅ Database Migrations\n")
		if m.Err == nil {
			b.WriteString("⏳ Functions\n")
		} else {
			b.WriteString("❌ Functions\n")
		}

	case StatusRunning:
		b.WriteString("✅ Schema\n")
		b.WriteString("✅ Database Migrations\n")
		b.WriteString("✅ Functions\n")
	}

	if len(m.MigrationChanges) > 0 {
		b.WriteString("\n")
		b.WriteString(colors.Heading("Schema changes:").String())
		b.WriteString("\n")
		for _, ch := range m.MigrationChanges {
			b.WriteString(" - ")
			switch ch.Type {
			case migrations.ChangeTypeAdded:
				b.WriteString(colors.Green(ch.Type).String())
			case migrations.ChangeTypeRemoved:
				b.WriteString(colors.Red(ch.Type).String())
			case migrations.ChangeTypeModified:
				b.WriteString(colors.Black(ch.Type).String())
			}
			b.WriteString(" ")
			b.WriteString(ch.Model)
			if ch.Field != "" {
				b.WriteString(fmt.Sprintf(".%s", ch.Field))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if m.Status == StatusRunning {
		if len(m.Schema.Apis) == 0 {
			b.WriteString(colors.Yellow("\n - Your schema doesn't have any API's defined in it").String())
		}

		for _, api := range m.Schema.Apis {
			b.WriteString("\n")
			b.WriteString(api.Name)
			b.WriteString(colors.White(" endpoints:").String())
			endpoints := [][]string{
				{"graphiql", "GraphiQL Playground"},
				{"graphql", "GraphQL"},
				{"json", "JSON"},
				{"rpc", "JSON-RPC"},
			}
			for _, values := range endpoints {
				b.WriteString("\n")
				b.WriteString(" - ")
				b.WriteString(colors.Blue(fmt.Sprintf("http://localhost:%s/%s/%s", m.Port, strings.ToLower(api.Name), values[0])).Highlight().String())
				b.WriteString(colors.White(fmt.Sprintf(" (%s)", values[1])).String())
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(colors.White(" - press ").String())
	b.WriteString("q")
	b.WriteString(colors.White(" to quit").String())
	b.WriteString("\n")

	return b.String()
}

func renderError(m *Model) string {
	b := strings.Builder{}

	switch m.Status {
	case StatusCheckingDependencies:
		incorrectNodeVersionErr := &node.IncorrectNodeVersionError{}
		if errors.As(m.Err, &incorrectNodeVersionErr) {
			b.WriteString(fmt.Sprintf("❌ You have Node %s installed but the minimum required is %s", incorrectNodeVersionErr.Current, incorrectNodeVersionErr.Minimum))
		} else if errors.Is(m.Err, &node.NodeNotFoundError{}) {
			b.WriteString("❌ Node is not installed or the executable's location is not added to $PATH")
		} else {
			b.WriteString("❌ There is an issue with your dependencies:\n\n")
			b.WriteString(m.Err.Error())
		}
	case StatusInitialized:
		b.WriteString("❌ There was an error initialising the Keel project:\n\n")
		b.WriteString(m.Err.Error())
	case StatusSetupDatabase:
		b.WriteString("❌ There was an error starting the database:\n\n")
		b.WriteString(m.Err.Error())

	case StatusSetupFunctions:
		npmInstallErr := &node.NpmInstallError{}
		if errors.As(m.Err, &npmInstallErr) {
			b.WriteString("❌ There was an error installing function dependencies:\n\n")
			b.WriteString(npmInstallErr.Output)
		} else {
			b.WriteString("❌ There was an error setting up your project:\n\n")
			b.WriteString(m.Err.Error())
		}

	case StatusLoadSchema:
		validationErrors := &errorhandling.ValidationErrors{}
		configErrors := &config.ConfigErrors{}

		switch {
		case errors.As(m.Err, &validationErrors):
			b.WriteString("❌ The following errors were found in your schema files:\n\n")
			s := validationErrors.ToAnnotatedSchema(m.SchemaFiles)
			b.WriteString(s)
		case errors.As(m.Err, &configErrors):
			b.WriteString("❌ The following errors were found in your ")
			b.WriteString(colors.Yellow("keelconfig.yaml").String())
			b.WriteString(" file:\n\n")
			for _, v := range configErrors.Errors {
				b.WriteString(" - ")
				b.WriteString(colors.Red(v.Message).String())
				b.WriteString("\n")
			}
		case m.Err == schema.ErrNoSchemaFiles:
			b.WriteString("❌ No Keel schema files found in: ")
			b.WriteString(colors.White(m.ProjectDir).String())
		default:
			b.WriteString("❌ There was an error loading your schema:\n\n")
			b.WriteString(m.Err.Error())
		}

	case StatusRunMigrations:
		b.WriteString("❌ There was an error updating your database schema:\n\n")

		dbErr := &db.DbError{}
		if errors.As(m.Err, &dbErr) {
			b.WriteString(colors.Red("Error: ").String())
			b.WriteString(colors.Red(dbErr.Message).String())
			b.WriteString(colors.Red(fmt.Sprintf(" (SQLSTATE Code: %s)", dbErr.PgErrCode)).String())
			b.WriteString("\n")

			if dbErr.Table != "" {
				b.WriteString(colors.Red("Table: ").String())
				b.WriteString(colors.Red(dbErr.Table).String())
				b.WriteString("\n")
			}

			if dbErr.Columns != nil && len(dbErr.Columns) > 0 {
				b.WriteString(colors.Red("Column(s): ").String())
				b.WriteString(colors.Red(strings.Join(dbErr.Columns, ", ")).String())
				b.WriteString("\n")
			}
		} else {
			b.WriteString(colors.Red(m.Err.Error()).String())
		}

	case StatusUpdateFunctions:
		tscError := &TypeScriptError{}
		if errors.As(m.Err, &tscError) && tscError.Output != "" {
			if strings.Contains(tscError.Output, "No inputs were found in config file") {
				b.WriteString(fmt.Sprintf("❌ Your functions/ folder is empty. Please run %s", colors.Cyan("keel generate").String()))
			} else {
				b.WriteString("❌ We found the following errors in your function code:\n\n")
				b.WriteString(tscError.Output)
			}

		} else {
			b.WriteString("❌ There was an error running your functions:\n\n")
			b.WriteString(m.Err.Error())
		}

	case StatusStartingFunctions:
		startFunctionsError := &StartFunctionsError{}
		b.WriteString("❌ There was an error running your functions:\n\n")
		b.WriteString(m.Err.Error())
		if errors.As(m.Err, &startFunctionsError) && startFunctionsError.Output != "" {
			b.WriteString("\n\n")
			b.WriteString(startFunctionsError.Output)
		}
	default:
		b.WriteString("❌ Oh no, looks like something went wrong:\n\n")
		b.WriteString(m.Err.Error())
	}

	b.WriteString("\n")
	return b.String()
}

func renderFunctionLog(log *FunctionLog) string {
	b := strings.Builder{}
	b.WriteString(colors.Yellow("[Functions]").String())
	b.WriteString(" ")
	b.WriteString(log.Value)

	return b.String()
}

func renderRequestLog(request *RuntimeRequest) string {
	b := strings.Builder{}

	b.WriteString(colors.Cyan("[Request]").String())
	b.WriteString(" ")
	b.WriteString(colors.White(request.Method).String())
	b.WriteString(" ")
	b.WriteString(request.Path)

	return b.String()
}

func RenderSecrets(secrets map[string]string) string {
	var rows []table.Row
	var keys []string
	for k := range secrets {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		rows = append(rows, table.Row{k, secrets[k]})
	}

	columns := []table.Column{
		{Title: "Name", Width: 50},
		{Title: "Value", Width: 50},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(keys)),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.NoColor{}).
		Bold(false)
	s.Cell = s.Cell.
		Foreground(colors.HighlightWhiteBright)

	t.SetStyles(s)

	secretsStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder())

	return secretsStyle.Render(t.View()) + "\n"
}

func RenderError(message error) error {
	return errors.New(colors.Red(message.Error()).Highlight().String())
}

func RenderSuccess(message string) {
	fmt.Println(colors.Green(message).Highlight().String())
}

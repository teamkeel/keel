package program

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func renderRun(m *Model) string {
	b := strings.Builder{}
	b.WriteString("Running Keel app in directory: ")
	b.WriteString(White(m.ProjectDir).Base())
	b.WriteString("\n")

	if m.DatabaseConnInfo != nil {
		b.WriteString("Connect to your database using: ")
		b.WriteString(Yellow(m.DatabaseConnInfo.String()).Base())
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
		b.WriteString(Heading("Schema changes:").Base())
		for _, ch := range m.MigrationChanges {
			b.WriteString("\n")
			b.WriteString(" - ")
			switch ch.Type {
			case migrations.ChangeTypeAdded:
				b.WriteString(Red(ch.Type).Base())
			case migrations.ChangeTypeRemoved:
				b.WriteString(Red(ch.Type).Base())
			case migrations.ChangeTypeModified:
				b.WriteString(Black(ch.Type).Base())
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
			b.WriteString(Yellow("\n - Your schema doesn't have any API's defined in it").Base())
		}

		for _, api := range m.Schema.Apis {
			b.WriteString("\n")
			b.WriteString(api.Name)
			b.WriteString(White(" endpoints:").Base())
			endpoints := [][]string{
				{"graphiql", "GraphiQL Playground"},
				{"graphql", "GraphQL"},
				{"json", "JSON"},
				{"rpc", "JSON-RPC"},
			}
			for _, values := range endpoints {
				b.WriteString("\n")
				b.WriteString(" - ")
				b.WriteString(Magenta(fmt.Sprintf("http://localhost:%s/%s/%s", m.Port, strings.ToLower(api.Name), values[0])).Base())
				b.WriteString(White(fmt.Sprintf(" (%s)", values[1])).Base())
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(White(" - press ").Base())
	b.WriteString("q")
	b.WriteString(White(" to quit").Base())
	b.WriteString("\n")

	return b.String()
}

func renderError(m *Model) string {
	b := strings.Builder{}

	switch m.Status {
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
			b.WriteString(Yellow("keelconfig.yaml").Base())
			b.WriteString(" file:\n\n")
			for _, v := range configErrors.Errors {
				b.WriteString(" - ")
				b.WriteString(Red(v.Message).Base())
				b.WriteString("\n")
			}
		case m.Err == schema.ErrNoSchemaFiles:
			b.WriteString("❌ No Keel schema files found in: ")
			b.WriteString(White(m.ProjectDir).Base())
		default:
			b.WriteString("❌ There was an error loading your schema:\n\n")
			b.WriteString(m.Err.Error())
		}

	case StatusRunMigrations:
		dbErr := &db.DbError{}

		b.WriteString("❌ There was an error updating your database schema:\n\n")
		b.WriteString("  ")
		if errors.As(m.Err, &dbErr) {
			b.WriteString("column ")
			b.WriteString(Red(dbErr.Column).Base())
			b.WriteString(": ")
			b.WriteString(Red(dbErr.Error()).Base())
		} else {
			b.WriteString(Red(m.Err.Error()).Base())
		}

	case StatusUpdateFunctions:
		tscError := &TypeScriptError{}
		if errors.As(m.Err, &tscError) && tscError.Output != "" {
			b.WriteString("❌ We found the following errors in your function code:\n\n")
			b.WriteString(tscError.Output)
		} else {
			b.WriteString("❌ There was an error running your functions:\n\n")
			b.WriteString(m.Err.Error())
		}

	case StatusStartingFunctions:
		startFunctionsError := &StartFunctionsError{}
		b.WriteString("❌ There was an error running your functions\n\n")
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

func renderLog(requests []*RuntimeRequest, functionLogs []*FunctionLog) string {
	b := strings.Builder{}

	b.WriteString(Heading("Log:").Highlight())
	b.WriteString("\n")

	type log struct {
		t     time.Time
		value string
	}
	logs := []*log{}

	for _, r := range requests {
		b := strings.Builder{}
		b.WriteString(Yellow("[Request]").Base())
		b.WriteString(" ")
		b.WriteString(White(r.Method).Base())
		b.WriteString(" ")
		b.WriteString(r.Path)
		logs = append(logs, &log{
			t:     r.Time,
			value: b.String(),
		})
	}

	for _, r := range functionLogs {
		b := strings.Builder{}
		b.WriteString(Yellow("[Functions]").Base())
		b.WriteString(" ")
		b.WriteString(r.Value)
		logs = append(logs, &log{
			t:     r.Time,
			value: b.String(),
		})
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].t.Before(logs[j].t)
	})

	for _, log := range logs {
		b.WriteString(log.value)
		b.WriteString("\n")
	}

	return b.String()
}

func renderSecrets(secrets map[string]string) string {
	var rows []table.Row
	var keys []string
	for k := range secrets {
		keys = append(keys, k)
	}

	sort.Sort(sort.StringSlice(keys))

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
		Foreground(highlightWhiteBright)

	t.SetStyles(s)

	secretsStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder())

	return secretsStyle.Render(t.View()) + "\n"
}

package program

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/go-wordwrap"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

func renderRun(m *Model) string {
	b := strings.Builder{}
	if m.Status == StatusQuitting {
		b.WriteString("Goodbye 👋")
		return b.String() + "\n"
	}

	b.WriteString("\n")

	if m.Status == StatusSeedCompleted {
		if len(m.SeededFiles) == 1 {
			b.WriteString("✅ Seeded data from 1 file\n")
		} else {
			b.WriteString(fmt.Sprintf("✅ Seeded data from %d files\n", len(m.SeededFiles)))
		}
		return b.String()
	}

	if m.Status == StatusSnapshotCompleted {
		b.WriteString("✅ Database snapshot saved to ./seed/snapshot.sql\n")
		return b.String()
	}

	if m.Mode == ModeRun {
		w := NewViewWriter(m.width)

		w.WriteWrapped("Running Keel app in directory: " + colors.White(m.ProjectDir).String())
		w.Newline()

		if m.DatabaseConnInfo != nil {
			connStr := fmt.Sprintf("psql -Atx \"%s\"", m.DatabaseConnInfo.String())
			w.WriteWithPrefix("Connect to your database using: ", colors.Cyan(connStr).Highlight().String())
			w.Newline()
		}

		if m.LatestVersion != nil {
			currentVersion, _ := semver.NewVersion(runtime.GetVersion())
			if currentVersion != nil && currentVersion.LessThan(m.LatestVersion) {
				w.Newline()
				w.WriteWrapped(colors.Red(fmt.Sprintf("There is a new version of Keel available. Please update to v%s by running this command:", m.LatestVersion.String())).String())
				w.Newline()
				w.Write(colors.White("$ ").String())
				w.Write(colors.Yellow("npm install -g keel").String())
				w.Newline()
			}
		}

		b.WriteString(w.String())
		b.WriteString("\n")
	}

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

	case StatusSeedData:
		b.WriteString("✅ Schema\n")
		b.WriteString("✅ Database Migrations\n")
		if m.Err == nil {
			b.WriteString("⏳ Seeding data...\n")
		} else {
			b.WriteString("❌ Seed data\n")
		}

	case StatusUpdateFunctions, StatusStartingFunctions:
		b.WriteString("✅ Schema\n")
		b.WriteString("✅ Database Migrations\n")

		if len(m.SeededFiles) > 0 {
			if len(m.SeededFiles) == 1 {
				b.WriteString("✅ Seeded data from 1 file\n")
			} else {
				b.WriteString(fmt.Sprintf("✅ Seeded data from %d files\n", len(m.SeededFiles)))
			}
		}

		if m.Err == nil {
			b.WriteString("⏳ Functions\n")
		} else {
			b.WriteString("❌ Functions\n")
		}

	case StatusRunning:
		b.WriteString("✅ Schema\n")
		b.WriteString("✅ Database Migrations\n")

		if len(m.SeededFiles) > 0 {
			if len(m.SeededFiles) == 1 {
				b.WriteString("✅ Seeded data from 1 file\n")
			} else {
				b.WriteString(fmt.Sprintf("✅ Seeded data from %d files\n", len(m.SeededFiles)))
			}
		}

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
	}

	if m.Status == StatusRunning {
		if len(m.Schema.GetApis()) == 0 {
			b.WriteString(colors.Yellow("\n - Your schema doesn't have any API's defined in it\n").String())
		}

		b.WriteString("\n")
		b.WriteString("Local development console: ")
		b.WriteString(colors.Blue("https://console.keel.so/local").Highlight().String())
		b.WriteString("\n")

		for _, api := range m.Schema.GetApis() {
			b.WriteString("\n")
			b.WriteString(api.GetName())
			b.WriteString(colors.White(" endpoints:").String())

			endpoints := [][]string{
				{"graphql", "GraphQL"},
				{"json", "JSON"},
				{"rpc", "JSON-RPC"},
			}

			for _, values := range endpoints {
				b.WriteString("\n")
				b.WriteString(" - ")
				if m.CustomHostname == "" {
					b.WriteString(colors.Blue(fmt.Sprintf("http://localhost:%s/%s/%s", m.Port, strings.ToLower(api.GetName()), values[0])).Highlight().String())
				} else {
					b.WriteString(colors.Blue(fmt.Sprintf("%s/%s/%s", m.CustomHostname, strings.ToLower(api.GetName()), values[0])).Highlight().String())
				}
				b.WriteString(colors.White(fmt.Sprintf(" (%s)", values[1])).String())
			}
			b.WriteString("\n")
		}
	}

	// Display recent function logs and runtime requests when running
	if m.Status == StatusRunning && m.Mode == ModeRun {
		const maxLogsToDisplay = 20

		// Display recent function logs with wrapping
		startIdx := 0
		if len(m.FunctionsLog) > maxLogsToDisplay {
			startIdx = len(m.FunctionsLog) - maxLogsToDisplay
		}
		for i := startIdx; i < len(m.FunctionsLog); i++ {
			b.WriteString(renderFunctionLogWrapped(m.FunctionsLog[i], m.width))
			b.WriteString("\n")
		}

		// Display recent runtime requests with wrapping
		startIdx = 0
		if len(m.RuntimeRequests) > maxLogsToDisplay {
			startIdx = len(m.RuntimeRequests) - maxLogsToDisplay
		}
		for i := startIdx; i < len(m.RuntimeRequests); i++ {
			b.WriteString(renderRequestLogWrapped(m.RuntimeRequests[i], m.width))
			b.WriteString("\n")
		}
	}

	if m.Mode == ModeRun {
		b.WriteString("\n")
		b.WriteString(colors.White("Press ").String())
		b.WriteString("q")
		b.WriteString(colors.White(" to quit").String())
		b.WriteString("\n")
	}

	return b.String()
}

func renderError(m *Model) string {
	w := NewViewWriter(m.width)

	switch m.Status {
	case StatusCheckingDependencies:
		incorrectNodeVersionErr := &node.IncorrectNodeVersionError{}
		if errors.As(m.Err, &incorrectNodeVersionErr) {
			w.WriteWrappedF("❌ You have Node %s installed but the minimum required is %s",
				incorrectNodeVersionErr.Current, incorrectNodeVersionErr.Minimum)
		} else if errors.Is(m.Err, &node.NodeNotFoundError{}) {
			w.WriteWrapped("❌ Node is not installed or the executable's location is not added to $PATH")
		} else {
			w.Write("❌ There is an issue with your dependencies:\n\n")
			w.WriteWrapped(m.Err.Error())
		}
	case StatusInitialized:
		w.Write("❌ There was an error initialising the Keel project:\n\n")
		w.WriteWrapped(m.Err.Error())
	case StatusSetupDatabase:
		w.Write("❌ There was an error starting the database:\n\n")
		w.WriteWrapped(m.Err.Error())

	case StatusSetupFunctions:
		npmInstallErr := &node.NpmInstallError{}
		if errors.As(m.Err, &npmInstallErr) {
			w.Write("❌ There was an error installing function dependencies:\n\n")
			w.WriteWrapped(npmInstallErr.Output)
		} else {
			w.Write("❌ There was an error setting up your project:\n\n")
			w.WriteWrapped(m.Err.Error())
		}

	case StatusLoadSchema:
		validationErrors := &errorhandling.ValidationErrors{}
		configErrors := &config.ConfigErrors{}

		switch {
		case errors.As(m.Err, &validationErrors):
			w.Write("❌ The following errors were found in your schema files:\n\n")
			s := validationErrors.ErrorsToAnnotatedSchema(m.SchemaFiles)
			w.Write(s)
		case errors.As(m.Err, &configErrors):
			w.Write("❌ The following errors were found in your ")
			w.Write(colors.Yellow("keelconfig.yaml").String())
			w.Write(" file:\n\n")
			for _, v := range configErrors.Errors {
				w.WriteWithPrefix(" - ", colors.Red(v.Message).String())
				w.Newline()
			}
		case m.Err == schema.ErrNoSchemaFiles:
			w.WriteWrapped("❌ No Keel schema files found in: " + colors.White(m.ProjectDir).String())
		default:
			w.Write("❌ There was an error loading your schema:\n\n")
			w.WriteWrapped(m.Err.Error())
		}

	case StatusRunMigrations:
		w.Write("❌ There was an error updating your database schema:\n\n")

		dbErr := &db.DbError{}
		if errors.As(m.Err, &dbErr) {
			w.WriteWrapped(colors.Red("Error: ").String() + colors.Red(dbErr.Message).String() +
				colors.Red(fmt.Sprintf(" (SQLSTATE Code: %s)", dbErr.PgErrCode)).String())
			w.Newline()

			if dbErr.Table != "" {
				w.WriteWrapped(colors.Red("Table: ").String() + colors.Red(dbErr.Table).String())
				w.Newline()
			}

			if len(dbErr.Columns) > 0 {
				w.WriteWrapped(colors.Red("Column(s): ").String() + colors.Red(strings.Join(dbErr.Columns, ", ")).String())
				w.Newline()
			}
		} else {
			w.WriteWrapped(colors.Red(m.Err.Error()).String())
		}

	case StatusUpdateFunctions:
		tscError := &TypeScriptError{}
		if errors.As(m.Err, &tscError) && tscError.Output != "" {
			if strings.Contains(tscError.Output, "No inputs were found in config file") {
				w.WriteWrappedF("❌ Your functions/ folder is empty. Please run %s", colors.Cyan("keel generate").String())
			} else {
				w.Write("❌ We found the following errors in your function code:\n\n")
				w.WriteWrapped(tscError.Output)
			}
		} else {
			w.Write("❌ There was an error running your functions:\n\n")
			w.WriteWrapped(m.Err.Error())
		}

	case StatusStartingFunctions:
		startFunctionsError := &StartFunctionsError{}
		w.Write("❌ There was an error running your functions:\n\n")
		w.WriteWrapped(m.Err.Error())
		if errors.As(m.Err, &startFunctionsError) && startFunctionsError.Output != "" {
			w.Write("\n\n")
			w.WriteWrapped(startFunctionsError.Output)
		}
	case StatusErrorStartingServers:
		w.Write("❌ There was an error starting the local servers:\n\n")
		w.WriteWrapped(m.Err.Error())

	default:
		w.Write("❌ Oh no, looks like something went wrong:\n\n")
		w.WriteWrapped(m.Err.Error())
	}

	w.Newline()
	return w.String()
}

// ViewWriter is a helper for writing text with automatic wrapping
// based on terminal width. This improves DX by not requiring width
// to be passed on every call.
type ViewWriter struct {
	width   int
	builder *strings.Builder
}

// NewViewWriter creates a new ViewWriter with the given terminal width
func NewViewWriter(width int) *ViewWriter {
	if width <= 0 {
		width = 80
	}
	return &ViewWriter{
		width:   width,
		builder: &strings.Builder{},
	}
}

// Write writes a string directly without wrapping
func (w *ViewWriter) Write(s string) {
	w.builder.WriteString(s)
}

// Writef writes a formatted string directly without wrapping
func (w *ViewWriter) Writef(format string, args ...interface{}) {
	w.builder.WriteString(fmt.Sprintf(format, args...))
}

// WriteWrapped writes text with smart wrapping (word wrap + hard wrap for long words)
// This handles both normal text and long paths/URLs without spaces
func (w *ViewWriter) WriteWrapped(text string) {
	wrapped := smartWrapText(text, w.width)
	w.builder.WriteString(wrapped)
}

// WriteWrappedF writes formatted text with smart wrapping
func (w *ViewWriter) WriteWrappedF(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	w.WriteWrapped(text)
}

// WriteWrappedHard writes text with hard wrapping (breaks anywhere, not just spaces)
// Useful when you want to force break at exact width regardless of spaces
func (w *ViewWriter) WriteWrappedHard(text string) {
	wrapped := hardWrapText(text, w.width)
	w.builder.WriteString(wrapped)
}

// WriteWithPrefix writes text with a prefix, wrapping and indenting continuation lines
// Uses smart wrapping to handle both regular text and long paths/URLs
func (w *ViewWriter) WriteWithPrefix(prefix, text string) {
	prefixLen := len(prefix)
	wrapped := smartWrapText(text, w.width-prefixLen)
	lines := strings.Split(wrapped, "\n")

	if len(lines) == 0 {
		return
	}

	// First line with prefix
	w.builder.WriteString(prefix)
	w.builder.WriteString(lines[0])

	// Subsequent lines with indentation
	if len(lines) > 1 {
		indent := strings.Repeat(" ", prefixLen)
		for i := 1; i < len(lines); i++ {
			w.builder.WriteString("\n")
			w.builder.WriteString(indent)
			w.builder.WriteString(lines[i])
		}
	}
}

// Newline writes a newline character
func (w *ViewWriter) Newline() {
	w.builder.WriteString("\n")
}

// String returns the accumulated string
func (w *ViewWriter) String() string {
	return w.builder.String()
}

// wrapText wraps text to fit within the given width
func wrapText(text string, width int) string {
	if width <= 0 {
		width = 80
	}
	// Reserve some space for prefix/styling
	wrapWidth := uint(width - 2)
	if wrapWidth < 20 {
		wrapWidth = 20
	}
	return wordwrap.WrapString(text, wrapWidth)
}

// hardWrapText wraps text by breaking at exact width, even within words
// This is useful for long paths, URLs, or content without spaces
func hardWrapText(text string, width int) string {
	if width <= 0 {
		width = 80
	}
	// For hard wrapping, allow any width (even small ones for paths)
	// Just ensure it's at least 10 to avoid unreasonable wrapping
	if width < 10 {
		width = 10
	}

	if len(text) <= width {
		return text
	}

	var result strings.Builder
	remaining := text

	for len(remaining) > 0 {
		if len(remaining) <= width {
			result.WriteString(remaining)
			break
		}

		// Take width characters
		chunk := remaining[:width]
		remaining = remaining[width:]

		result.WriteString(chunk)
		if len(remaining) > 0 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// smartWrapText intelligently wraps text, using word wrapping when possible
// but falling back to hard wrapping for long words without spaces
func smartWrapText(text string, width int) string {
	if width <= 0 {
		width = 80
	}
	if width < 20 {
		width = 20
	}

	// First try word wrapping
	wrapped := wordwrap.WrapString(text, uint(width-2))

	// Check if any line is still too long (no spaces to break on)
	lines := strings.Split(wrapped, "\n")
	var result strings.Builder

	for i, line := range lines {
		if len(line) <= width {
			result.WriteString(line)
		} else {
			// Line is too long, apply hard wrapping to this line
			result.WriteString(hardWrapText(line, width))
		}

		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func renderFunctionLog(log *FunctionLog) string {
	b := strings.Builder{}
	b.WriteString(colors.Yellow("[Functions]").String())
	b.WriteString(" ")
	b.WriteString(log.Value)

	return b.String()
}

func renderFunctionLogWrapped(log *FunctionLog, width int) string {
	w := NewViewWriter(width)
	prefix := colors.Yellow("[Functions]").String() + " "
	w.WriteWithPrefix(prefix, log.Value)
	return w.String()
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

func renderRequestLogWrapped(request *RuntimeRequest, width int) string {
	w := NewViewWriter(width)
	prefix := colors.Cyan("[Request]").String() + " " + colors.White(request.Method).String() + " "
	w.WriteWithPrefix(prefix, request.Path)
	return w.String()
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

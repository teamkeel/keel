package program

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/reader"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

var (
	red     = color.New(color.FgRed)
	green   = color.New(color.FgHiGreen, color.Faint)
	blue    = color.New(color.FgHiBlue)
	yellow  = color.New(color.FgYellow)
	gray    = color.New(color.FgWhite, color.Faint)
	heading = color.New(color.FgWhite, color.Underline, color.Bold)
)

func renderRun(m *Model) string {
	b := strings.Builder{}
	b.WriteString("Running Keel app in directory: ")
	b.WriteString(gray.Sprint(m.ProjectDir))
	b.WriteString("\n")

	if m.DatabaseConnInfo != nil {
		b.WriteString("Connect to your database using: ")
		b.WriteString(yellow.Sprint(m.DatabaseConnInfo.String()))
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
		b.WriteString(heading.Sprint("Schema changes:\n"))
		for _, ch := range m.MigrationChanges {
			b.WriteString(" - ")
			switch ch.Type {
			case migrations.ChangeTypeAdded:
				b.WriteString(green.Sprint(ch.Type))
			case migrations.ChangeTypeRemoved:
				b.WriteString(red.Sprint(ch.Type))
			case migrations.ChangeTypeModified:
				b.WriteString(yellow.Sprint(ch.Type))
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
			b.WriteString(yellow.Sprint("\n - Your schema doesn't have any API's defined in it"))
		}

		for _, api := range m.Schema.Apis {
			b.WriteString("\n")
			b.WriteString(api.Name)
			b.WriteString(gray.Sprint(" endpoints:\n"))
			endpoints := [][]string{
				{"graphiql", "GraphiQL Playground"},
				{"graphql", "GraphQL"},
				{"json", "JSON"},
				{"rpc", "JSON-RPC"},
			}
			for _, values := range endpoints {
				b.WriteString(" - ")
				b.WriteString(blue.Sprintf("http://localhost:%s/%s/%s", m.Port, strings.ToLower(api.Name), values[0]))
				b.WriteString(gray.Sprintf(" (%s)\n", values[1]))
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(gray.Sprint(" - press "))
	b.WriteString("q")
	b.WriteString(gray.Sprint(" to quit"))
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
			s := renderSchemaValidationErrors(validationErrors, m.SchemaFiles)
			b.WriteString(s)
		case errors.As(m.Err, &configErrors):
			b.WriteString("❌ The following errors were found in your ")
			b.WriteString(yellow.Sprint("keelconfig.yaml"))
			b.WriteString(" file:\n\n")
			for _, v := range configErrors.Errors {
				b.WriteString(" - ")
				b.WriteString(red.Sprintf(v.Message))
				b.WriteString("\n")
			}
		case m.Err == schema.ErrNoSchemaFiles:
			b.WriteString("❌ No Keel schema files found in: ")
			b.WriteString(gray.Sprint(m.ProjectDir))
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
			b.WriteString(red.Sprint(dbErr.Column))
			b.WriteString(": ")
			b.WriteString(red.Sprint(dbErr.Error()))
		} else {
			b.WriteString(red.Sprintf(m.Err.Error()))
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

	b.WriteString(heading.Sprint("Log:"))
	b.WriteString("\n")

	type log struct {
		t     time.Time
		value string
	}
	logs := []*log{}

	for _, r := range requests {
		b := strings.Builder{}
		b.WriteString(yellow.Sprint("[Request]"))
		b.WriteString(" ")
		b.WriteString(gray.Sprintf(r.Method))
		b.WriteString(" ")
		b.WriteString(r.Path)
		logs = append(logs, &log{
			t:     r.Time,
			value: b.String(),
		})
	}

	for _, r := range functionLogs {
		b := strings.Builder{}
		b.WriteString(yellow.Sprint("[Functions]"))
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

// renderSchemaValidationErrors formats the given errors by pointing to the relevant line
// in the source file that produced the error
func renderSchemaValidationErrors(verrs *errorhandling.ValidationErrors, sources []reader.SchemaFile) string {

	// Number of lines of the source code to render before and after the line with the error
	bufferLines := 3

	// How much to indent the entire result e.g. every line is indented this much
	indent := 2

	result := strings.Repeat(" ", indent)
	newLine := func() {
		result += "\n" + strings.Repeat(" ", indent)
	}

	for _, err := range verrs.Errors {
		// Assumption here is that the error is on one line
		errorLine := err.Pos.Line

		startSourceLine := errorLine - bufferLines
		endSourceLine := errorLine + bufferLines

		// This produces a format string like "%3s| " which we use to render the gutter
		// The number after the "%" is the width, which is documented as:
		//   > For most values, width is the minimum number of runes to output,
		//   > padding the formatted form with spaces if necessary.
		gutterFmt := "%" + fmt.Sprintf("%d", len(fmt.Sprintf("%d", endSourceLine))) + "s| "

		var source string
		for _, s := range sources {
			if s.FileName == err.Pos.Filename {
				source = s.Contents
				break
			}
		}

		result += green.Sprintf(gutterFmt, " ")
		result += green.Sprint(err.Pos.Filename)
		newLine()

		// not sure this can happen, but just in case we'll handle it
		if source == "" {
			result += err.Message
			newLine()
			continue
		}

		lines := strings.Split(source, "\n")

		for lineIndex, line := range lines {

			// If this line is outside of the buffer we can drop it
			if (lineIndex+1) < (startSourceLine) || (lineIndex+1) > (endSourceLine) {
				continue
			}

			// Render line numbers in gutter
			result += gray.Sprintf(gutterFmt, fmt.Sprintf("%d", lineIndex+1))

			// If the error line doesn't match the currently enumerated line
			// then we can render the whole line without any colorization
			if (lineIndex + 1) != errorLine {
				result += gray.Sprintf("%s", line)
				newLine()
				continue
			}

			chars := strings.Split(line, "")

			// Enumerate over the characters in the line
			for charIdx, char := range chars {

				// Check if the character index is less than or greater than the corresponding start and end column
				// If so, then render the char without any colorization
				if (charIdx+1) < err.Pos.Column || (charIdx+1) > err.EndPos.Column-1 {
					result += char
					continue
				}

				result += red.Sprint(char)
			}

			newLine()

			// Underline the token that caused the error
			result += gray.Sprintf(gutterFmt, "")
			result += strings.Repeat(" ", err.Pos.Column-1)
			tokenLength := err.EndPos.Column - err.Pos.Column
			for i := 0; i < tokenLength; i++ {
				if i == tokenLength/2 {
					result += yellow.Sprint("\u252C")
				} else {
					result += yellow.Sprint("\u2500")
				}
			}
			newLine()

			// Render the down arrow
			result += gray.Sprintf(gutterFmt, "")
			result += strings.Repeat(" ", err.Pos.Column-1)
			result += strings.Repeat(" ", (err.EndPos.Column-err.Pos.Column)/2)
			result += yellow.Sprint("\u2570")
			result += yellow.Sprint("\u2500")

			// Render the message
			result += fmt.Sprintf(" %s %s", yellow.Sprint(err.ErrorDetails.Message), red.Sprintf("(%s)", err.Code))
			newLine()

			// Render the hint
			if err.ErrorDetails.Hint != "" {
				result += gray.Sprintf(gutterFmt, "")
				result += strings.Repeat(" ", err.Pos.Column-1)
				result += strings.Repeat(" ", (err.EndPos.Column-err.Pos.Column)/2)
				// Line up hint with the error message above (taking into account unicode arrows)
				result += strings.Repeat(" ", 3)
				result += yellow.Sprint(err.ErrorDetails.Hint)
				newLine()
			}
		}

		newLine()
	}

	return result
}

package program

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/reader"
)

// These statuses are specific to the generate cmd
// The statuses are ordered by when they are executed.
const (
	StatusCheckEligibility = iota
	StatusNotGenerated
	StatusBootstrapping
	StatusNpmInstalling
	StatusGeneratingNodePackages
	StatusScaffolding
	StatusGenerated
)

type GenerateModel struct {
	// The directory of the Keel project
	ProjectDir string

	// If set then @teamkeel/* npm packages will be installed
	// from this path, rather than NPM.
	NodePackagesPath string

	Environment string

	Status int

	Err         error
	Schema      *proto.Schema
	SchemaFiles []reader.SchemaFile
	Secrets     map[string]string
	Config      *config.ProjectConfig

	GeneratedFiles codegen.GeneratedFiles

	generateCh chan tea.Msg

	generateOutput []*GenerateMsg

	npmInstallSpinner spinner.Model

	// Maintain the current dimensions of the user's terminal
	width  int
	height int
}

func (m *GenerateModel) Init() tea.Cmd {
	m.generateCh = make(chan tea.Msg, 1)

	m.Status = StatusLoadSchema

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	m.npmInstallSpinner = s

	return tea.Batch(LoadSchema(m.ProjectDir, "development"), m.npmInstallSpinner.Tick)
}

func (m *GenerateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Status = StatusQuitting
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// This msg is sent once on program start
		// and then subsequently every time the user
		// resizes their terminal window.
		m.width = msg.Width
		m.height = msg.Height

		return m, nil
	case LoadSchemaMsg:
		m.Err = msg.Err

		if msg.Err != nil {
			m.Status = StatusNotGenerated

			return m, tea.Quit
		}

		m.Schema = msg.Schema
		m.SchemaFiles = msg.SchemaFiles
		m.Config = msg.Config
		m.Secrets = msg.Secrets

		return m, tea.Batch(
			NextMsgCommand(m.generateCh),
			Generate(m.ProjectDir, m.Schema, m.NodePackagesPath, m.generateCh),
		)
	case GenerateMsg:
		m.generateOutput = append(m.generateOutput, &msg)
		m.Status = msg.Status
		m.GeneratedFiles = append(m.GeneratedFiles, msg.GeneratedFiles...)
		m.Err = msg.Err

		if msg.Err != nil {
			m.Status = StatusNotGenerated
			return m, tea.Quit
		}

		switch m.Status {
		case StatusGenerated, StatusNotGenerated:
			// if generation has finished successfully or generation was aborted
			// due to there being no tests / functions in the schema we can quit now.
			return m, tea.Quit
		default:
			// otherwise, keep reading from the channel
			return m, NextMsgCommand(m.generateCh)
		}
	default:
		var cmd tea.Cmd
		m.npmInstallSpinner, cmd = m.npmInstallSpinner.Update(msg)

		return m, cmd
	}

	return m, nil
}

func (m *GenerateModel) View() string {
	b := strings.Builder{}

	// lipgloss will automatically wrap any text based on the current dimensions of the user's term.
	s := lipgloss.
		NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height)

	b.WriteString(m.renderGenerate())

	return s.Render(b.String() + "\n")
}

func (m *GenerateModel) renderGenerate() string {
	b := strings.Builder{}

	if m.Status == StatusNotGenerated {
		if m.Err != nil {
			b.WriteString(colors.Red(fmt.Sprintf("Error: %s", m.Err.Error())).String())
		} else {
			b.WriteString(colors.Blue("⚠️  Not required").String())
			b.WriteString("\n\n")
			b.WriteString(colors.Gray("In order to use the ").String())
			b.WriteString(colors.Cyan("generate").String())
			b.WriteString(colors.Gray(" command, define some functions in your schema, or write some tests.").String())
			b.WriteString("\n\n")
			b.WriteString(colors.Gray("For more information, visit https://docs.keel.so/local-environment").String())
		}

		b.WriteString("\n")

		return b.String()
	}

	if m.Status >= StatusBootstrapping {
		b.WriteString(fmt.Sprintf("%s\n", colors.Cyan("🥾 Bootstrapping..").String()))
		relevant := m.filterLogsByStage(StatusBootstrapping)

		for _, msg := range relevant {
			b.WriteString(fmt.Sprintf("%s\n", colors.Gray(msg.Log).String()))
		}
	}

	if m.Status >= StatusNpmInstalling {
		b.WriteString("\n")
		b.WriteString(colors.Cyan("🏃 Installing dependencies..").String())

		if m.Status == StatusNpmInstalling {
			b.WriteString("\n")
			b.WriteString(m.npmInstallSpinner.View())
		} else {
			relevant := m.filterLogsByStage(StatusNpmInstalling)

			for _, msg := range relevant {
				b.WriteString(fmt.Sprintf("%s\n", colors.Gray(msg.Log).String()))
			}
		}

	}

	if m.Status >= StatusGeneratingNodePackages {
		b.WriteString("\n")

		b.WriteString(fmt.Sprintf("%s\n", colors.Cyan("📦 Generating dynamic packages..").String()))
		relevant := m.filterLogsByStage(StatusGeneratingNodePackages)

		for _, msg := range relevant {
			b.WriteString(fmt.Sprintf("%s\n", colors.Gray(msg.Log).String()))
		}
	}

	if m.Status >= StatusScaffolding {
		b.WriteString("\n")

		b.WriteString(fmt.Sprintf("%s\n", colors.Cyan("👷 Scaffolding missing functions..").String()))

		relevant := m.filterLogsByStage(StatusScaffolding)

		for _, msg := range relevant {
			b.WriteString(fmt.Sprintf("%s\n", colors.Gray(msg.Log).String()))
		}
	}
	if m.Status >= StatusGenerated {
		b.WriteString("\n")

		b.WriteString(fmt.Sprintf("%s\n\n", colors.Green("✅ All done!").String()))

		if len(m.GeneratedFiles) > 0 {
			b.WriteString(fmt.Sprintf("%s\n\n", colors.Gray("The following functions were generated:").String()))

			// output scaffolded file names with the function name highlighted in cyan
			for _, generatedFile := range m.GeneratedFiles {
				functionName := strings.Split(filepath.Base(generatedFile.Path), ".")[0]
				parts := strings.Split(generatedFile.Path, "/")
				prePath := filepath.Join(parts[0 : len(parts)-1]...)

				b.WriteString(
					colors.Gray(
						fmt.Sprintf("- %s/%s%s", prePath, colors.Cyan(functionName).String(), colors.Gray(".ts").String()),
					).Highlight().String(),
				)
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

func (m *GenerateModel) filterLogsByStage(status int) []*GenerateMsg {
	relevant := lo.Filter(m.generateOutput, func(o *GenerateMsg, _ int) bool {
		return o.Status == status
	})

	start := 5

	// handle case where there are less than 5 logs for the given status
	// and therefore tailing the last 5 doesn't make sense (and will result in a out of range err)
	if len(relevant) < start {
		start = 0
	} else {
		start = len(relevant) - start
	}

	relevant = lo.Slice(relevant, start, len(relevant))

	return relevant
}

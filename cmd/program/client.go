package program

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/reader"
)

const (
	StatusGeneratingClient = iota
	StatusNotGenerated
	StatusGenerated
)

type GenerateClientModel struct {
	// The directory of the Keel project
	ProjectDir string
	Package    bool
	Watch      bool
	OutputDir  string
	ApiName    string

	Status int

	Err         error
	Schema      *proto.Schema
	SchemaFiles []*reader.SchemaFile
	Secrets     map[string]string
	Config      *config.ProjectConfig

	GeneratedFiles codegen.GeneratedFiles

	generateCh chan tea.Msg
	watcherCh  chan tea.Msg

	generateOutput []*GenerateClientMsg

	// Maintain the current dimensions of the user's terminal
	width  int
	height int
}

func (m *GenerateClientModel) Init() tea.Cmd {
	m.generateCh = make(chan tea.Msg, 1)
	m.watcherCh = make(chan tea.Msg, 1)

	m.Status = StatusLoadSchema

	cmds := []tea.Cmd{
		LoadSchema(m.ProjectDir, "development"),
	}

	filter := []string{
		".keel",
	}

	if m.Watch {
		cmds = append(
			cmds,
			StartWatcher(m.ProjectDir, m.watcherCh, filter),
			NextMsgCommand(m.watcherCh),
		)
	}

	return tea.Batch(
		cmds...,
	)
}

func (m *GenerateClientModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if !m.Watch {
				return m, tea.Quit
			}
		}

		m.Schema = msg.Schema
		m.SchemaFiles = msg.SchemaFiles
		m.Config = msg.Config
		m.Secrets = msg.Secrets

		return m, tea.Batch(
			NextMsgCommand(m.generateCh),
			GenerateClient(m.ProjectDir, m.Schema, m.ApiName, m.OutputDir, m.Package, m.generateCh),
		)
	case GenerateClientMsg:
		m.generateOutput = append(m.generateOutput, &msg)
		m.Status = msg.Status
		// m.GeneratedFiles = append(m.GeneratedFiles, msg.GeneratedFiles...)
		m.GeneratedFiles = msg.GeneratedFiles
		m.Err = msg.Err

		if msg.Err != nil {
			m.Status = StatusNotGenerated
			if !m.Watch {
				return m, tea.Quit
			}
		}

		switch m.Status {
		case StatusGenerated, StatusNotGenerated:
			// if generation has finished successfully or generation was aborted
			// due to there being no tests / functions in the schema we can quit now.

			if !m.Watch {
				return m, tea.Quit
			}

		default:
			// otherwise, keep reading from the channel
			return m, NextMsgCommand(m.generateCh)
		}
	case WatcherMsg:
		m.Err = msg.Err
		m.Status = StatusGeneratingClient

		// If the watcher errors then probably best to exit
		if m.Err != nil {
			return m, tea.Quit
		}

		return m, tea.Batch(
			NextMsgCommand(m.watcherCh),
			LoadSchema(m.ProjectDir, "development"),
		)

	default:
		var cmd tea.Cmd

		return m, cmd
	}

	return m, nil
}

func (m *GenerateClientModel) View() string {
	b := strings.Builder{}

	// lipgloss will automatically wrap any text based on the current dimensions of the user's term.
	s := lipgloss.
		NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height)

	b.WriteString(m.renderGenerate())

	return s.Render(b.String() + "\n")
}

func (m *GenerateClientModel) renderGenerate() string {
	b := strings.Builder{}

	if m.Status == StatusNotGenerated {
		if m.Err != nil {
			b.WriteString(colors.Red(fmt.Sprintf("Error: %s", m.Err.Error())).String())
		} else {
			b.WriteString(colors.Blue("⚠️  Client could not be generated").String())
			b.WriteString("\n\n")
			b.WriteString(colors.Gray("Ensure you have a valid Keel schema in this directory and an API defined").String())
		}

		return b.String()
	}

	if m.Status >= StatusGeneratingClient {
		b.WriteString(fmt.Sprintf("%s\n", colors.Cyan("📦 Generating client SDK..").String()))
	}

	if m.Status >= StatusGenerated {

		b.WriteString(fmt.Sprintf("%s\n\n", colors.Green("✅ All done!").String()))

		if len(m.GeneratedFiles) > 0 {
			b.WriteString(fmt.Sprintf("%s\n\n", colors.Gray("The following files were generated:").String()))

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

	if m.Watch {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%s\n", colors.Gray("Watching for updates. Press ctrl+c or q to exit").String()))
	}

	return b.String()
}

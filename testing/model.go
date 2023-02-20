package testing

// This file heavily utilizes the terminal UI framework (TUI) Bubbletea: https://github.com/charmbracelet/bubbletea
// Take some time to read the documentation. Their discord server is very helpful for getting help.
// Bubbletea is a model-view framework. The Update() method on a model is responsible for receiving "messages" of
// different types (e.g window resize events, custom events from outside of the Bubbletea program that send new data etc),
// and updating the internal model state
// Once an Update() has finished, then the View() method is called to repaint the whole of the Terminal UI.
// Because the whole CLI output is repainted with every update to the state, you need to be careful to ensure you do not
// cause the UI to jump if you are mutating data in the update or if you loop over a data structure non deterministically.

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/muesli/termenv"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/testing/viewport"
)

const (
	StatusPass      = "pass"
	StatusFail      = "fail"
	StatusException = "exception"
)

type EventStatus string

const (
	EventStatusPending  EventStatus = "pending"
	EventStatusComplete EventStatus = "complete"
)

type Event struct {
	EventStatus EventStatus

	Result *TestResult
	Meta   *TestCase
}

type TestCase struct {
	TestName string `json:"name"`
	FilePath string `json:"filePath"`
}

type TestResult struct {
	TestCase

	Status   string          `json:"status"`
	Expected json.RawMessage `json:"expected,omitempty"`
	Actual   json.RawMessage `json:"actual,omitempty"`
	Err      json.RawMessage `json:"err,omitempty"`
}

// A Bubbletea model is responsible for maintaining the CLI state
// More about models at https://github.com/charmbracelet/bubbletea#the-model
type Model struct {
	spinner  spinner.Model
	viewport viewport.Model
	color    termenv.Profile

	onQuit func()

	builder  strings.Builder
	ready    bool
	tests    []*UITestCase
	cursor   int
	updating bool

	finished bool

	passedCount    int
	failedCount    int
	completedTests int

	Err error
}

type (
	errMsg struct {
		err error
	}
)

func NewModel(onQuit func()) *Model {
	s := spinner.NewModel()
	s.Spinner = spinner.Spinner{
		Frames: []string{"...", "·..", ".·.", "..·", "..."},
		FPS:    time.Second / 8,
	}
	s.Style.Foreground(lipgloss.Color("41"))

	return &Model{
		color:   termenv.ColorProfile(),
		spinner: s,
		tests:   make([]*UITestCase, 0),
		onQuit:  onQuit,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	allCmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case errMsg:
		// Handle any error message types raised and quit the tea program
		m.Err = msg.err

		return m, tea.Quit
	case []*Event:
		// Event is a catchall struct which notifies of changes to test results.
		// There are two scenarios:
		// 1. All of the test names are collected first before each test is run - these are "Pending" type events (all of the pending events come in at once in one array)
		// 2. When each test is executed, a "completed" event is pushed to this code with details of whether the test succeeeded / failed / error-ed
		if msg[0].EventStatus == EventStatusPending {
			m.ready = true
			for _, evt := range msg {
				s := spinner.New()
				s.Spinner = spinner.Spinner{
					Frames: []string{"....", "·...", ".·..", "..·.", "...·"},
					FPS:    time.Second / 8,
				}
				allCmds = append(allCmds, s.Tick)

				m.tests = append(m.tests, &UITestCase{
					TestName: evt.Meta.TestName,
					FilePath: evt.Meta.FilePath,
					spinner:  s,
				})
			}
		} else {
			evt := msg[0]
			for i, test := range m.tests {

				caseMatch := strings.EqualFold(test.TestName, evt.Result.TestName)

				if caseMatch {
					m.tests[i].Completed = true
					m.tests[i].StatusStr = evt.Result.Status

					if i >= m.viewport.Height-3 {
						m.viewport.YOffset++
					}
					m.completedTests++

					switch test.StatusStr {
					case StatusPass:
						m.passedCount++
					case StatusFail:
						m.tests[i].Actual = evt.Result.Actual
						m.tests[i].Expected = evt.Result.Expected
						m.failedCount++
					case StatusException:
						e := &JsError{}

						err := json.Unmarshal(evt.Result.Err, &e)

						if err != nil {
							continue
						}

						m.tests[i].Err = e
						m.failedCount++
					}

					allCmds = append(allCmds, m.tests[i].spinner.Tick)
				}
			}

			if m.completedTests == len(m.tests) {
				summary := m.failedTestSummary(m.failedTests())
				failedTestHeight := len(strings.Split(summary, "\n"))
				newHeight := m.viewport.TotalLineCount() + failedTestHeight

				m.viewport.Height = newHeight
				m.viewport.YOffset = 0
				m.finished = true
			}
			if m.completedTests == len(m.tests) && m.finished {
				allCmds = append(allCmds, tea.DisableMouse)
			}
		}

		return m, tea.Batch(allCmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.onQuit()
			return m, tea.Quit
		case "down", "j":
			if !m.updating {
				m.cursor++
				m.fixCursor()
				m.fixViewport(false)
			}
		case "up", "k":
			if !m.updating {
				m.cursor--
				m.fixCursor()
				m.fixViewport(false)
			}
		case "pgup", "u":
			if !m.updating {
				m.viewport.LineUp(1)
				m.fixViewport(true)
			}
		case "pgdown", "d":
			if !m.updating {
				m.viewport.LineDown(1)
				m.fixViewport(true)
			}
		}
	case spinner.TickMsg:
		// A tick message is Bubbletea's internal way of progressing the animation of loading spinners
		// Every time a previous tick message comes in, we want to append the new tick cmd to the update
		// so that spinners continue to spin
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		allCmds = append(allCmds, cmd)
		for i, t := range m.tests {
			s, cmd := t.spinner.Update(msg)

			m.tests[i].spinner = s

			allCmds = append(allCmds, cmd)
		}

		return m, tea.Batch(allCmds...)
	case tea.WindowSizeMsg:
		// The window size message event is triggered once when the program begins - the height and width that are sent are
		// the height and width of the terminal window
		// The message event is triggered with subsequent resizing of the window

		// We construct/mutate the height and width of our internal viewport based on the width and height of the window
		// As more tests results are reported, elsewhere we also want to increase the height of the viewport so we can display
		// all of the tests
		if !m.ready {
			m.viewport = viewport.Model{
				Width:  msg.Width,
				Height: msg.Height - 2,
			}
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 2
			m.fixViewport(true)
		}
	}

	viewport, cmd := m.viewport.Update(msg)

	m.viewport = viewport
	allCmds = append(allCmds, cmd)

	return m, tea.Batch(allCmds...)
}

func (m *Model) View() string {
	var header, body string
	if len(m.tests) < 1 {
		header = colors.Cyan(fmt.Sprint("Preparing tests" + m.spinner.View())).Base()
	} else {
		header = colors.White(fmt.Sprintf("Running %d tests", len(m.tests))).Bold().Base()
		m.viewport.SetContent(m.content())
		body = m.viewport.View()
	}

	return fmt.Sprintf("%s\n%s", header, body)
}

// Content is where all of the tests are (re)rendered
func (m *Model) content() string {
	defer m.builder.Reset()

	m.builder.WriteString("\n")

	// We need to compute the longest test name so we can add spacer between the progress counter and the names of the tests
	longestTestName := lo.MaxBy(m.tests, func(t *UITestCase, max *UITestCase) bool {
		return len(t.TestName) > len(max.TestName)
	})

	for i, test := range m.tests {
		if test.Completed {
			var c *colors.Colors

			if test.StatusStr == "pass" {
				c = colors.Green("")
			} else {
				c = colors.Red("")
			}

			m.builder.WriteString(
				fmt.Sprintf("%s  %s\n", c.UpdateText(fmt.Sprintf(" %s ", prettyStatusStr(test))).Base(), test.TestName),
			)
		} else if i > 0 && m.tests[i-1].Completed || i == 0 && !m.tests[i].Completed {
			m.builder.WriteString(
				fmt.Sprintf(
					" %s   %s%s(%d/%d)\n",
					test.spinner.View(),
					test.TestName,
					strings.Repeat(" ", len(longestTestName.TestName)-len(m.tests[i].TestName)+4),
					m.completedTests+1,
					len(m.tests),
				),
			)
		} else {
			m.builder.WriteString(
				fmt.Sprintf("        %s\n", colors.Green(fmt.Sprint(test.TestName)).Highlight()),
			)
		}
	}

	// Print the summary
	// e.g 5 passed · 1 failed · 6 total
	if m.finished {
		dialogBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#fff")).
			Width(50).
			Height(1).
			MarginTop(1).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			AlignHorizontal(lipgloss.Center).
			BorderBottom(true).
			MarginRight(2)

		m.builder.WriteString(
			dialogBoxStyle.Render(
				fmt.Sprintf("%s · %s · %s", colors.Green(fmt.Sprintf("%d passed", m.passedCount)).Base(), colors.Red(fmt.Sprintf("%d failed", m.failedCount)).Base(), colors.White(fmt.Sprintf("%d total", len(m.tests))).Base()),
			),
		)

		// failures
		if len(m.failedTests()) > 0 {
			m.builder.WriteString(m.failedTestSummary(m.failedTests()))
		}

		m.builder.WriteString("\n")
	}

	return m.builder.String()
}

func (m *Model) failedTests() []*UITestCase {
	return lo.Filter(m.tests, func(t *UITestCase, _ int) bool {
		return t.StatusStr != StatusPass
	})
}

func (m *Model) failedTestSummary(failedTests []*UITestCase) (s string) {
	dialogBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("1")).
		MarginTop(1).
		BorderTop(true).
		Width(30).
		BorderLeft(true).
		BorderRight(true).
		AlignHorizontal(lipgloss.Center).
		BorderBottom(true)

	s += "\n"
	s += fmt.Sprintf("%s\n\n", colors.Red(fmt.Sprintf("%d failed tests:", len(failedTests))).Highlight())

	withinBox := ""

	for _, failedTest := range failedTests {
		withinBox += fmt.Sprintf("%s %s", colors.White(fmt.Sprintf(" %s ", prettyStatusStr(failedTest))).Background(colors.StatusRedBright).Base(), failedTest.TestName)
		switch failedTest.StatusStr {
		case StatusFail:

			withinBox += lipgloss.JoinHorizontal(
				lipgloss.Center,
				dialogBoxStyle.Render(colors.Red(fmt.Sprintf("%s", failedTest.Expected)).Highlight()),
				dialogBoxStyle.Render(colors.Red(fmt.Sprintf("%s", failedTest.Actual)).Highlight()),
			)
			withinBox += "\n"
			labelsBox := lipgloss.NewStyle().
				Width(30).
				AlignHorizontal(lipgloss.Left).
				MarginRight(5)

			withinBox += lipgloss.JoinHorizontal(
				lipgloss.Center,
				labelsBox.Render(colors.Red("Expected").Highlight()),
				labelsBox.Render(colors.Red("Actual").Highlight()),
			)
			withinBox += "\n\n"
		case StatusException:
			withinBox += dialogBoxStyle.Width(m.viewport.Width - 5).Render(
				fmt.Sprintf(
					"%s\n%s",
					colors.Red(fmt.Sprint(failedTest.Err.Message)).Highlight(),
					colors.Red(fmt.Sprint(failedTest.Err.Stack)).Highlight(),
				),
			)

			withinBox += "\n\n"
		}
	}

	s += withinBox

	s += "\n\n"
	return s
}

func (m *Model) fixCursor() {
	if m.cursor > len(m.tests)-1 {
		m.cursor = len(m.tests) - 1
	} else if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) fixViewport(moveCursor bool) {
	top := m.viewport.YOffset
	bottom := m.viewport.Height + m.viewport.YOffset - 1

	if moveCursor {
		if m.cursor < top {
			m.cursor = top
		} else if m.cursor > bottom {
			m.cursor = bottom
		}
		return
	}

	if m.cursor < top {
		m.viewport.LineUp(top - m.cursor)
	} else if m.cursor > bottom {
		m.viewport.LineDown(m.cursor - bottom)
	}
}

type JsError struct {
	Message string `json:"message"`
	Stack   string `json:"stack"`
}

type UITestCase struct {
	TestName string
	FilePath string

	Completed bool
	StatusStr string

	Actual   any
	Expected any

	Err *JsError

	spinner spinner.Model
}

func prettyStatusStr(t *UITestCase) string {
	switch t.StatusStr {
	case StatusException:
		return "ERR "
	default:
		return strings.ToUpper(t.StatusStr)
	}
}

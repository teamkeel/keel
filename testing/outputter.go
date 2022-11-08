package testing

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Outputter struct {
	program *tea.Program

	OnQuit func()
}

func NewOutputter(workingDir string, onQuit func()) *Outputter {
	m := NewModel(onQuit)
	p := tea.NewProgram(m, tea.WithMouseCellMotion())

	return &Outputter{
		program: p,
		OnQuit:  onQuit,
	}
}

func (o *Outputter) Start() {
	go func() {
		if err := o.program.Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}()
}

func (o *Outputter) End() {
	o.program.ReleaseTerminal()
	// below is necessary to avoid weird control chars being
	// rendered to term after quit when using mousewheel
	o.program.DisableMouseAllMotion()
}

func (o *Outputter) Push(evts []*Event) {
	o.program.Send(evts)
}

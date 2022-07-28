package functions

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

type Default struct {
}

type StaticAnalysisResult struct {
	Default Default `json:"default"`
}

type StaticAnalyser struct {
	Workdir string

	Result StaticAnalysisResult
}

func NewStaticAnalyser(workdir string) *StaticAnalyser {
	return &StaticAnalyser{
		Workdir: workdir,
	}
}

func (a *StaticAnalyser) Analyse() error {
	cliPath := filepath.Join(a.Workdir, "node_modules", "@teamkeel", "client", "dist", "cli.js")

	cmd := exec.Command("node", cliPath, a.Workdir)

	o, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Print(string(o))
		return err
	}

	return nil
}

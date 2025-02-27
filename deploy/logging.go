package deploy

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/teamkeel/keel/colors"
)

var (
	IconCross = colors.Red("✘").String()
	IconTick  = colors.Green("✔").String()
	IconPipe  = colors.Yellow("|").String()
	LogIndent = "  "
)

func log(format string, a ...any) {
	fmt.Printf(LogIndent+format+"\n", a...)
}

func heading(v string) {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingLeft(1).
		PaddingRight(1)

	log("\n%s%s  ", LogIndent, style.Render(v))
}

func orange(format string, a ...any) string {
	v := fmt.Sprintf(format, a...)
	return colors.Orange(v).String()
}

func gray(format string, a ...any) string {
	v := fmt.Sprintf(format, a...)
	return colors.Gray(v).String()
}

func green(format string, a ...any) string {
	v := fmt.Sprintf(format, a...)
	return colors.Green(v).String()
}

func red(format string, a ...any) string {
	v := fmt.Sprintf(format, a...)
	return colors.Red(v).String()
}

type Timing struct {
	t time.Time
}

func NewTiming() *Timing {
	return &Timing{t: time.Now()}
}

// Since returns a gray string containing the duration since it was last called or the Timing struct was created
func (t *Timing) Since() string {
	since := time.Since(t.t)
	since = since - (since % time.Millisecond)
	v := gray("(%s)", since.String())
	t.t = time.Now()
	return v
}

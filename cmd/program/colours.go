package program

import "github.com/charmbracelet/lipgloss"

var (
	statusGreen            = lipgloss.AdaptiveColor{Light: "#086c08", Dark: "#009900"}
	statusGreenBright      = lipgloss.AdaptiveColor{Light: "#086c08", Dark: "#009900"}
	statusYellow           = lipgloss.AdaptiveColor{Light: "#999900", Dark: "#999900"}
	statusYellowBright     = lipgloss.AdaptiveColor{Light: "#FFFF00", Dark: "#FFFF00"}
	statusRed              = lipgloss.AdaptiveColor{Light: "#990000", Dark: "#990000"}
	statusRedBright        = lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}
	statusBlue             = lipgloss.AdaptiveColor{Light: "#000099", Dark: "#000099"}
	statusBlueBright       = lipgloss.AdaptiveColor{Light: "#0000FF", Dark: "#0000FF"}
	highlightBlack         = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
	highlightBlackBright   = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#666666"}
	highlightWhite         = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
	highlightWhiteBright   = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
	highlightMagenta       = lipgloss.AdaptiveColor{Light: "#990099", Dark: "#FF00FF"}
	highlightMagentaBright = lipgloss.AdaptiveColor{Light: "#990099", Dark: "#FF00FF"}
	highlightCyan          = lipgloss.AdaptiveColor{Light: "#009999", Dark: "#00FFFF"}
	highlightCyanBright    = lipgloss.AdaptiveColor{Light: "#009999", Dark: "#00FFFF"}
)

type Colors struct {
	BaseColor      lipgloss.AdaptiveColor
	HighlightColor lipgloss.AdaptiveColor
	text           lipgloss.Style
}

func (c *Colors) Base() string {
	return c.text.Foreground(c.BaseColor).String()
}

func (c *Colors) Highlight() string {
	return c.text.Foreground(c.HighlightColor).String()
}

func setText(text string) lipgloss.Style {
	return lipgloss.NewStyle().
		SetString(text).Width(len(text))
}

func Red(text string) *Colors {
	return &Colors{
		BaseColor:      statusRed,
		HighlightColor: statusRedBright,
		text:           setText(text),
	}
}

func Green(text string) *Colors {
	return &Colors{
		BaseColor:      statusGreen,
		HighlightColor: statusGreenBright,
		text:           setText(text),
	}
}

func Yellow(text string) *Colors {
	return &Colors{
		BaseColor:      statusYellow,
		HighlightColor: statusYellowBright,
		text:           setText(text),
	}
}

func Blue(text string) *Colors {
	return &Colors{
		BaseColor:      statusBlue,
		HighlightColor: statusBlueBright,
		text:           setText(text),
	}
}

func Magenta(text string) *Colors {
	return &Colors{
		BaseColor:      highlightMagenta,
		HighlightColor: highlightMagentaBright,
		text:           setText(text),
	}
}

func Cyan(text string) *Colors {
	return &Colors{
		BaseColor:      highlightCyan,
		HighlightColor: highlightCyanBright,
		text:           setText(text),
	}
}

func White(text string) *Colors {
	return &Colors{
		BaseColor:      highlightWhite,
		HighlightColor: highlightWhiteBright,
		text:           setText(text),
	}
}

func Black(text string) *Colors {
	return &Colors{
		BaseColor:      highlightBlack,
		HighlightColor: highlightBlackBright,
		text:           setText(text),
	}
}

func Gray(text string) *Colors {
	return &Colors{
		BaseColor:      highlightWhite,
		HighlightColor: highlightWhiteBright,
		text:           setText(text),
	}
}

func Heading(text string) *Colors {
	return &Colors{
		BaseColor:      highlightBlack,
		HighlightColor: highlightBlackBright,
		text:           setText(text),
	}
}

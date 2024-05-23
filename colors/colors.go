package colors

import "github.com/charmbracelet/lipgloss"

var (
	StatusGreen            = lipgloss.AdaptiveColor{Light: "#086c08", Dark: "#009900"}
	StatusGreenBright      = lipgloss.AdaptiveColor{Light: "#086c08", Dark: "#00FF00"}
	StatusYellow           = lipgloss.AdaptiveColor{Light: "#999900", Dark: "#999900"}
	StatusYellowBright     = lipgloss.AdaptiveColor{Light: "#B3B300", Dark: "#FFFF00"}
	StatusRed              = lipgloss.AdaptiveColor{Light: "#990000", Dark: "#CC0000"}
	StatusRedBright        = lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF0000"}
	StatusOrange           = lipgloss.AdaptiveColor{Light: "#CC7700", Dark: "#FFAA00"}
	StatusOrangeBright     = lipgloss.AdaptiveColor{Light: "#FFAA00", Dark: "#FFAA00"}
	StatusBlue             = lipgloss.AdaptiveColor{Light: "#000099", Dark: "#000099"}
	StatusBlueBright       = lipgloss.AdaptiveColor{Light: "#0000FF", Dark: "#6699ff"}
	HighlightBlack         = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
	HighlightBlackBright   = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#666666"}
	HighlightWhite         = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#999999"}
	HighlightWhiteBright   = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
	HighlightMagenta       = lipgloss.AdaptiveColor{Light: "#990099", Dark: "#FF00FF"}
	HighlightMagentaBright = lipgloss.AdaptiveColor{Light: "#990099", Dark: "#FF00FF"}
	HighlightCyan          = lipgloss.AdaptiveColor{Light: "#009999", Dark: "#00FFFF"}
	HighlightCyanBright    = lipgloss.AdaptiveColor{Light: "#009999", Dark: "#00FFFF"}
)

type Colors struct {
	BaseColor      lipgloss.AdaptiveColor
	HighlightColor lipgloss.AdaptiveColor
	text           lipgloss.Style
}

func (c *Colors) String() string {
	return c.text.String()
}

func (c *Colors) Highlight() *Colors {
	return &Colors{
		BaseColor:      c.BaseColor,
		HighlightColor: c.HighlightColor,
		text:           c.text.Foreground(c.HighlightColor),
	}
}

func (c *Colors) Bold() *Colors {
	return &Colors{
		BaseColor:      c.BaseColor,
		HighlightColor: c.HighlightColor,
		text:           c.text.Bold(true),
	}
}

func (c *Colors) Background(color lipgloss.AdaptiveColor) *Colors {
	return &Colors{
		BaseColor:      c.BaseColor,
		HighlightColor: c.HighlightColor,
		text:           c.text.Background(color),
	}
}

func (c *Colors) UpdateText(text string) *Colors {
	return &Colors{
		BaseColor:      c.BaseColor,
		HighlightColor: c.HighlightColor,
		text:           setText(text, c.BaseColor),
	}
}

func setText(text string, colour lipgloss.AdaptiveColor) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colour).
		SetString(text)
}

func Red(text string) *Colors {
	return &Colors{
		BaseColor:      StatusRed,
		HighlightColor: StatusRedBright,
		text:           setText(text, StatusRed),
	}
}

func Green(text string) *Colors {
	return &Colors{
		BaseColor:      StatusGreen,
		HighlightColor: StatusGreenBright,
		text:           setText(text, StatusGreen),
	}
}

func Yellow(text string) *Colors {
	return &Colors{
		BaseColor:      StatusYellow,
		HighlightColor: StatusYellowBright,
		text:           setText(text, StatusYellow),
	}
}

func Orange(text string) *Colors {
	return &Colors{
		BaseColor:      StatusOrange,
		HighlightColor: StatusOrangeBright,
		text:           setText(text, StatusOrange),
	}
}

func Blue(text string) *Colors {
	return &Colors{
		BaseColor:      StatusBlue,
		HighlightColor: StatusBlueBright,
		text:           setText(text, StatusBlue),
	}
}

func Magenta(text string) *Colors {
	return &Colors{
		BaseColor:      HighlightMagenta,
		HighlightColor: HighlightMagentaBright,
		text:           setText(text, HighlightMagenta),
	}
}

func Cyan(text string) *Colors {
	return &Colors{
		BaseColor:      HighlightCyan,
		HighlightColor: HighlightCyanBright,
		text:           setText(text, HighlightCyan),
	}
}

func White(text string) *Colors {
	return &Colors{
		BaseColor:      HighlightWhite,
		HighlightColor: HighlightWhiteBright,
		text:           setText(text, HighlightWhite),
	}
}

func Black(text string) *Colors {
	return &Colors{
		BaseColor:      HighlightBlack,
		HighlightColor: HighlightBlackBright,
		text:           setText(text, HighlightBlack),
	}
}

func Gray(text string) *Colors {
	return &Colors{
		BaseColor:      HighlightBlackBright,
		HighlightColor: HighlightWhite,
		text:           setText(text, HighlightBlackBright),
	}
}

func Heading(text string) *Colors {
	return &Colors{
		BaseColor:      HighlightBlack,
		HighlightColor: HighlightBlackBright,
		text:           setText(text, HighlightBlack).Underline(true),
	}
}

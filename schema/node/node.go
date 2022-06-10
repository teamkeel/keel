package node

import (
	"unicode/utf8"

	"github.com/alecthomas/participle/v2/lexer"
)

type ParserNode interface {
	GetPositionRange() (start lexer.Position, end lexer.Position)
}

type Node struct {
	Pos    lexer.Position
	EndPos lexer.Position
	Tokens []lexer.Token
}

// GetPositionRange returns a start and end position that correspond to Node
// The behaviour of start position is exactly the same as the Pos field that
// participle provides but the end position is calculated from the position of
// the last token in this node, which is more useful if you want to know where
// _this_ node starts and ends.
func (n Node) GetPositionRange() (start lexer.Position, end lexer.Position) {
	start.Column = n.Pos.Column
	start.Filename = n.Pos.Filename
	start.Line = n.Pos.Line
	start.Offset = n.Pos.Offset

	// This shouldn't really happen but just to be safe
	if len(n.Tokens) == 0 {
		return start, n.EndPos
	}

	lastToken := n.Tokens[len(n.Tokens)-1]
	endPos := lastToken.Pos

	tokenLength := utf8.RuneCountInString(lastToken.Value)

	end.Filename = endPos.Filename

	// assumption here is that a token can't span multiple lines, which
	// I'm pretty sure is true
	end.Line = endPos.Line

	// Update offset and column to reflect the end of last token
	// in this node
	end.Offset = endPos.Offset + tokenLength
	end.Column = endPos.Column + tokenLength

	return start, end
}

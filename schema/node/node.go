package node

import (
	"unicode/utf8"

	"github.com/alecthomas/participle/v2/lexer"
)

type ParserNode interface {
	GetPositionRange() (start lexer.Position, end lexer.Position)
	InRange(position Position) bool
	HasEndPosition() bool
	GetTokens() []lexer.Token
}

type Position struct {
	Filename string `json:"filename"`
	Column   int    `json:"column"`
	Line     int    `json:"line"`
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
	// Note: However, clearing out the tokens in some cases is useful when you want to render a substring
	// that cannot otherwise be easily tokenized by the lexer.
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

func (n Node) InRange(position Position) bool {
	line := position.Line
	column := position.Column

	hasEndPos := n.EndPos.Column != 0 && n.EndPos.Line != 0

	if hasEndPos {
		// line before
		if line < n.Pos.Line {
			return false
		}

		// line after
		if line > n.EndPos.Line {
			return false
		}

		// if the line in the editor is the same line as the start of the tokens
		if line == n.Pos.Line {
			// if the column is less than the start pos
			// then its not in range
			return column < n.Pos.Column
		}

		// if the position is in-between the range of start and end lines,
		// return true
		if line > n.Pos.Line && line < n.EndPos.Line {
			return true
		}

		// if the line is the same line as the last token
		// and the col is less than the column of the end of the last token
		// return true
		if line == n.Pos.Line && column <= n.EndPos.Column {
			return true
		}

		// otherwise the line is the same line,
		// but the column value is greater than the column
		// of the last token
		return false
	}

	// if there is no end pos, we can only compare the start
	if line == n.Pos.Line {
		return column >= n.Pos.Column
	} else if line < n.Pos.Line {
		return false
	}

	return true
}

// Due to the way we have structured our parser fields
// Collection blocks with no content such as fields {} return
// no nodes which means we cant purely rely on node.inRange check
// so it is necessary to examine the underlying tokens
// emitted by the parser
func BoundaryTokensInRange(position Position, start lexer.Token, end lexer.Token) bool {
	col := position.Column
	line := position.Line

	// line before
	if line < start.Pos.Line {
		return false
	}

	// line after
	if line > end.Pos.Line {
		return false
	}

	// if the line in the editor is the same line as the start of the tokens
	if line == start.Pos.Line {
		// if the column is less than the start pos
		// then its not in range
		return col < start.Pos.Column
	}

	// if the position is in-between the range of start and end lines,
	// return true
	if line > start.Pos.Line && line < end.Pos.Line {
		return true
	}

	// if the line is the same line as the last token
	// and the col is less than the column of the end of the last token
	// return true
	if line == end.Pos.Line && col <= end.Pos.Column {
		return true
	}

	// otherwise the line is the same line,
	// but the column value is greater than the column
	// of the last token
	return false
}

func (n Node) HasEndPosition() bool {
	// Nodes in parents and after where a syntax error have occurred will have these values below
	return n.EndPos.Filename != "" && n.EndPos.Column > 0 && n.EndPos.Line > 0
}

func (n Node) GetTokens() []lexer.Token {
	return n.Tokens
}

package actions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"github.com/teamkeel/keel/colors"
)

func Test_DebugVisitor(t *testing.T) {
	expression := `var1 == var2 && var3 || var4 == "hello"`
	expected := `
open
	open
		open
			var1 Equals var2
		close
		&&
		open
			var3
		close
	close
	||
	open
		var4 Equals hello
	close
close`

	res, err := RunCelVisitor(expression, DebugStringGenerator())
	assert.NoError(t, err)

	// diff := diffmatchpatch.New()
	// diffs := diff.DiffMain(expected, generator.String(), true)
	if !strings.Contains(expected, res) {
		//t.Errorf("generated code does not match expected:\n%s", diffs)
		t.Errorf("\nExpected:\n---------\n%s", expected)
		t.Errorf("\nActual:\n---------\n%s", res)
	}

}

// diffPrettyText is a port of the same function from the diffmatchpatch
// lib but with better handling of whitespace diffs (by using background colours)
func diffPrettyText(diffs []diffmatchpatch.Diff) string {
	var buff strings.Builder

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			if strings.TrimSpace(diff.Text) == "" {
				buff.WriteString(colors.Green(fmt.Sprint(diff.Text)).String())
			} else {
				buff.WriteString(colors.Green(fmt.Sprint(diff.Text)).Highlight().String())
			}
		case diffmatchpatch.DiffDelete:
			if strings.TrimSpace(diff.Text) == "" {
				buff.WriteString(colors.Red(diff.Text).String())
			} else {
				buff.WriteString(colors.Red(fmt.Sprint(diff.Text)).Highlight().String())
			}
		case diffmatchpatch.DiffEqual:
			buff.WriteString(diff.Text)
		}
	}

	return buff.String()
}

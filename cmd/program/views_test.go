package program

import (
	"strings"
	"testing"
)

// TestViewWriter_Basic tests basic write functionality
func TestViewWriter_Basic(t *testing.T) {
	w := NewViewWriter(80)
	w.Write("Hello")
	w.Write(" ")
	w.Write("World")

	result := w.String()
	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got %q", result)
	}
}

// TestViewWriter_Writef tests formatted writing
func TestViewWriter_Writef(t *testing.T) {
	w := NewViewWriter(80)
	w.Writef("Hello %s, you are %d years old", "Alice", 30)

	result := w.String()
	expected := "Hello Alice, you are 30 years old"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestViewWriter_WriteWrapped tests text wrapping
func TestViewWriter_WriteWrapped(t *testing.T) {
	w := NewViewWriter(40)
	longText := "This is a very long text that should wrap because it exceeds the terminal width"

	w.WriteWrapped(longText)
	result := w.String()

	// Should contain multiple lines
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Errorf("Expected text to wrap to multiple lines, got %d lines", len(lines))
	}

	// Each line should be within reasonable width
	for i, line := range lines {
		if len(line) > 42 { // Allow small margin
			t.Errorf("Line %d exceeds width: %d > 42", i, len(line))
		}
	}
}

// TestViewWriter_WriteWrappedF tests formatted wrapped writing
func TestViewWriter_WriteWrappedF(t *testing.T) {
	w := NewViewWriter(30)
	w.WriteWrappedF("Error: The file %s could not be found in directory %s", "config.yaml", "/very/long/path/to/directory")

	result := w.String()
	lines := strings.Split(result, "\n")

	if len(lines) < 2 {
		t.Errorf("Expected wrapped text to span multiple lines")
	}
}

// TestViewWriter_WriteWithPrefix tests prefix writing with indentation
func TestViewWriter_WriteWithPrefix(t *testing.T) {
	w := NewViewWriter(50)
	prefix := "[ERROR] "
	text := "This is a long error message that should wrap with proper indentation"

	w.WriteWithPrefix(prefix, text)
	result := w.String()

	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Logf("Result:\n%s", result)
		t.Error("Expected text to wrap to multiple lines")
		return
	}

	// First line should start with prefix
	if !strings.HasPrefix(lines[0], prefix) {
		t.Errorf("First line should start with prefix %q, got %q", prefix, lines[0])
	}

	// Second line should be indented
	if len(lines) > 1 {
		expectedIndent := strings.Repeat(" ", len(prefix))
		if !strings.HasPrefix(lines[1], expectedIndent) {
			t.Errorf("Second line should be indented with %d spaces, got %q", len(prefix), lines[1][:min(len(lines[1]), len(prefix))])
		}
	}
}

// TestViewWriter_Newline tests newline insertion
func TestViewWriter_Newline(t *testing.T) {
	w := NewViewWriter(80)
	w.Write("Line 1")
	w.Newline()
	w.Write("Line 2")

	result := w.String()
	expected := "Line 1\nLine 2"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestViewWriter_MixedUsage tests combining different write methods
func TestViewWriter_MixedUsage(t *testing.T) {
	w := NewViewWriter(60)

	w.Write("Header\n")
	w.WriteWrappedF("Error: %s", "Something went wrong with a very long explanation that needs wrapping")
	w.Newline()
	w.WriteWithPrefix("  > ", "Additional context that is also quite long and should wrap properly")

	result := w.String()

	if !strings.Contains(result, "Header") {
		t.Error("Expected result to contain 'Header'")
	}
	if !strings.Contains(result, "Error:") {
		t.Error("Expected result to contain 'Error:'")
	}
	if !strings.Contains(result, "  > ") {
		t.Error("Expected result to contain prefix '  > '")
	}
}

// TestViewWriter_ZeroWidth tests handling of zero/invalid width
func TestViewWriter_ZeroWidth(t *testing.T) {
	w := NewViewWriter(0) // Should default to 80
	w.WriteWrapped("Short text")

	result := w.String()
	if result != "Short text" {
		t.Errorf("Expected 'Short text', got %q", result)
	}
}

// TestViewWriter_EmptyString tests handling empty strings
func TestViewWriter_EmptyString(t *testing.T) {
	w := NewViewWriter(80)
	w.Write("")
	w.WriteWrapped("")

	result := w.String()
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

// TestSmartWrapText_LongPathWithoutSpaces tests wrapping of long paths
func TestSmartWrapText_LongPathWithoutSpaces(t *testing.T) {
	longPath := "/Users/tomfrew/Developer/playground/team_operate_test/node_modules/@teamkeel/functions-runtime/dist/index.cjs"
	width := 40

	result := smartWrapText(longPath, width)
	lines := strings.Split(result, "\n")

	// Should wrap into multiple lines
	if len(lines) < 2 {
		t.Errorf("Expected long path to wrap into multiple lines, got %d lines", len(lines))
	}

	// Each line should be within width
	for i, line := range lines {
		if len(line) > width+2 { // Small margin
			t.Errorf("Line %d exceeds width: %d > %d\nLine: %s", i, len(line), width, line)
		}
	}

	// The path should still be present in full
	combined := strings.Join(lines, "")
	if combined != longPath {
		t.Error("Path was modified during wrapping")
	}
}

// TestSmartWrapText_MixedContent tests smart wrapping with mixed spaces and long words
func TestSmartWrapText_MixedContent(t *testing.T) {
	text := "Error: Cannot find module '/very/long/path/to/some/module/that/has/no/spaces/file.js' in the current directory"
	width := 50

	result := smartWrapText(text, width)
	lines := strings.Split(result, "\n")

	// Should wrap into multiple lines
	if len(lines) < 2 {
		t.Logf("Result:\n%s", result)
		t.Error("Expected text to wrap into multiple lines")
	}

	// Each line should be reasonable length
	for i, line := range lines {
		if len(line) > width+5 { // Allow some margin
			t.Logf("Line %d: %q (len=%d)", i, line, len(line))
			t.Errorf("Line %d exceeds width significantly", i)
		}
	}
}

// TestHardWrapText tests hard wrapping at exact width
func TestHardWrapText(t *testing.T) {
	longText := "abcdefghijklmnopqrstuvwxyz0123456789"
	width := 10

	result := hardWrapText(longText, width)
	lines := strings.Split(result, "\n")

	// Should break at exactly 10 characters
	if len(lines) != 4 { // 36 chars / 10 = 3 full lines + 1 partial
		t.Errorf("Expected 4 lines, got %d", len(lines))
	}

	// First 3 lines should be exactly 10 chars
	for i := 0; i < 3; i++ {
		if len(lines[i]) != 10 {
			t.Errorf("Line %d should be 10 chars, got %d", i, len(lines[i]))
		}
	}

	// Reconstruct should match original
	combined := strings.Join(lines, "")
	if combined != longText {
		t.Error("Text was modified during hard wrapping")
	}
}

// TestViewWriter_WriteWrappedWithLongPath tests ViewWriter with long paths
func TestViewWriter_WriteWrappedWithLongPath(t *testing.T) {
	w := NewViewWriter(50)
	longPath := "/Users/username/very/long/path/to/some/deeply/nested/directory/structure/file.txt"

	w.WriteWrapped(longPath)
	result := w.String()

	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Error("Expected long path to wrap")
	}

	// Should preserve the full path
	combined := strings.Join(lines, "")
	if combined != longPath {
		t.Error("Path was modified")
	}
}

// TestViewWriter_WriteWithPrefixLongPath tests prefix with long path
func TestViewWriter_WriteWithPrefixLongPath(t *testing.T) {
	w := NewViewWriter(60)
	prefix := "Error: Cannot find module "
	longPath := "'/Users/tomfrew/Developer/test/node_modules/@teamkeel/functions/dist/index.js'"

	w.WriteWithPrefix(prefix, longPath)
	result := w.String()

	lines := strings.Split(result, "\n")

	// First line should have prefix
	if !strings.HasPrefix(lines[0], prefix) {
		t.Error("First line should have prefix")
	}

	// If wrapped, continuation lines should be indented
	if len(lines) > 1 {
		expectedIndent := strings.Repeat(" ", len(prefix))
		for i := 1; i < len(lines); i++ {
			if !strings.HasPrefix(lines[i], expectedIndent) {
				t.Errorf("Line %d should be indented", i)
			}
		}
	}
}

// min is a helper function for minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

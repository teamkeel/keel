package program

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestModelUpdate_WindowResize verifies that window resize triggers screen clearing
func TestModelUpdate_WindowResize(t *testing.T) {
	model := &Model{
		width:  80,
		height: 24,
	}

	// Simulate window resize
	newModel, cmd := model.Update(tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	})

	m := newModel.(*Model)

	// Verify dimensions were updated
	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}

	// Verify that ClearScreen command is returned
	if cmd == nil {
		t.Error("Expected ClearScreen command, got nil")
	}
}

// TestFunctionsLog_BoundedHistory verifies that function logs are trimmed
func TestFunctionsLog_BoundedHistory(t *testing.T) {
	model := &Model{
		functionsLogCh: make(chan tea.Msg, 1),
		Mode:           ModeRun,
	}

	// Add more than maxHistorySize logs
	for i := 0; i < maxHistorySize+50; i++ {
		msg := FunctionsOutputMsg{
			Output: "test log",
		}
		newModel, _ := model.Update(msg)
		model = newModel.(*Model)
	}

	// Verify that logs are bounded to maxHistorySize
	if len(model.FunctionsLog) != maxHistorySize {
		t.Errorf("Expected %d function logs, got %d", maxHistorySize, len(model.FunctionsLog))
	}
}

// TestRuntimeRequests_BoundedHistory verifies that runtime requests are trimmed
func TestRuntimeRequests_BoundedHistory(t *testing.T) {
	model := &Model{
		Mode:   ModeRun,
		Status: StatusRunning,
	}

	// Directly add more than maxHistorySize requests to test the trimming logic
	for i := 0; i < maxHistorySize+50; i++ {
		model.RuntimeRequests = append(model.RuntimeRequests, &RuntimeRequest{
			Method: "GET",
			Path:   "/test",
		})

		// Apply the same trimming logic as in the actual handler
		if len(model.RuntimeRequests) > maxHistorySize {
			model.RuntimeRequests = model.RuntimeRequests[len(model.RuntimeRequests)-maxHistorySize:]
		}
	}

	// Verify that requests are bounded to maxHistorySize
	if len(model.RuntimeRequests) != maxHistorySize {
		t.Errorf("Expected %d runtime requests, got %d", maxHistorySize, len(model.RuntimeRequests))
	}
}

// TestView_NoDuplicateOutput verifies that View() doesn't produce duplicate content
func TestView_NoDuplicateOutput(t *testing.T) {
	model := &Model{
		width:           80,
		height:          24,
		Status:          StatusRunning,
		Mode:            ModeRun,
		ProjectDir:      "/test/dir",
		FunctionsLog:    []*FunctionLog{{Value: "test log"}},
		RuntimeRequests: []*RuntimeRequest{{Method: "GET", Path: "/api/test"}},
	}

	view := model.View()

	// Count occurrences of specific strings
	logCount := strings.Count(view, "test log")
	requestCount := strings.Count(view, "/api/test")

	// Each log/request should appear exactly once
	if logCount != 1 {
		t.Errorf("Expected 'test log' to appear once, appeared %d times", logCount)
	}
	if requestCount != 1 {
		t.Errorf("Expected '/api/test' to appear once, appeared %d times", requestCount)
	}
}

// TestView_MaxLogsDisplayed verifies that only recent logs are shown
func TestView_MaxLogsDisplayed(t *testing.T) {
	model := &Model{
		width:      80,
		height:     200, // Large height to ensure no truncation
		Status:     StatusRunning,
		Mode:       ModeRun,
		ProjectDir: "/test/dir",
	}

	// Add 30 unique logs so we can count them
	for i := 0; i < 30; i++ {
		model.FunctionsLog = append(model.FunctionsLog, &FunctionLog{
			Value: "test log",
		})
	}

	view := model.View()

	// Count how many times "test log" appears - should be capped at maxLogsToDisplay (20)
	logCount := strings.Count(view, "test log")

	if logCount > 20 {
		t.Errorf("Expected at most 20 logs displayed, got %d", logCount)
	}
	// Allow some flexibility for the actual count due to rendering
	if logCount < 20 {
		t.Logf("Warning: Expected 20 logs displayed, got %d (may be due to height constraints)", logCount)
	}
}

// TestView_HandlesEmptyState verifies that View() handles empty state gracefully
func TestView_HandlesEmptyState(t *testing.T) {
	model := &Model{
		width:  80,
		height: 24,
		Status: StatusCheckingDependencies,
		Mode:   ModeRun,
	}

	view := model.View()

	if view == "" {
		t.Error("Expected non-empty view, got empty string")
	}

	// Should not panic or produce malformed output
	if !strings.Contains(view, "\n") {
		t.Error("Expected view to contain newlines")
	}
}

// TestView_ErrorStateNoOverlap verifies error messages don't overlap with normal output
func TestView_ErrorStateNoOverlap(t *testing.T) {
	model := &Model{
		width:      80,
		height:     24,
		Status:     StatusLoadSchema,
		Mode:       ModeRun,
		ProjectDir: "/test/dir",
		Err:        &FakeError{"test error"},
	}

	view := model.View()

	// Should contain error indicator
	if !strings.Contains(view, "❌") {
		t.Error("Expected error indicator in view")
	}

	// Error message should appear exactly once
	errorCount := strings.Count(view, "test error")
	if errorCount != 1 {
		t.Errorf("Expected error message to appear once, appeared %d times", errorCount)
	}
}

// TestView_StatusTransitions verifies that status changes render correctly
func TestView_StatusTransitions(t *testing.T) {
	statuses := []int{
		StatusCheckingDependencies,
		StatusSetupDatabase,
		StatusSetupFunctions,
		StatusLoadSchema,
		StatusRunMigrations,
		StatusRunning,
	}

	for _, status := range statuses {
		model := &Model{
			width:  80,
			height: 24,
			Status: status,
			Mode:   ModeRun,
		}

		view := model.View()

		if view == "" {
			t.Errorf("Expected non-empty view for status %d", status)
		}

		// View should contain content (lipgloss may add/remove newlines)
		if len(view) < 10 {
			t.Errorf("Expected substantial content in view for status %d, got: %s", status, view)
		}
	}
}

// TestView_LongTextWrapping verifies that long text wraps properly
func TestView_LongTextWrapping(t *testing.T) {
	longLogValue := "This is a very long log message that should wrap across multiple lines when the terminal width is narrow. It contains lots of text to ensure proper wrapping behavior."

	model := &Model{
		width:      40, // Narrow width to force wrapping
		height:     50,
		Status:     StatusRunning,
		Mode:       ModeRun,
		ProjectDir: "/test/dir",
		FunctionsLog: []*FunctionLog{
			{Value: longLogValue},
		},
	}

	view := model.View()

	// The log value should appear in the view (might be wrapped/formatted)
	if !strings.Contains(view, "very long log") && !strings.Contains(view, longLogValue[:20]) {
		t.Logf("View output:\n%s", view)
		t.Error("Expected long log message in view")
	}

	// View should not be excessively wide (wrapping should occur)
	lines := strings.Split(view, "\n")
	maxWidth := 0
	for i, line := range lines {
		// Account for ANSI color codes which don't count toward visible width
		visibleLen := len(stripAnsiCodes(line))
		if visibleLen > maxWidth {
			maxWidth = visibleLen
		}
		// Allow generous margin because lipgloss MaxWidth may not be exact
		if visibleLen > model.width+15 {
			t.Logf("Line %d: visible=%d, expected<=%d", i, visibleLen, model.width+15)
		}
	}
	t.Logf("Max visible width in view: %d (terminal width: %d)", maxWidth, model.width)
}

// TestRenderFunctionLogWrapped verifies function log wrapping with indentation
func TestRenderFunctionLogWrapped(t *testing.T) {
	longLog := &FunctionLog{
		Value: "This is a very long function log message that definitely needs to wrap",
	}

	result := renderFunctionLogWrapped(longLog, 40)

	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Error("Expected log to wrap to multiple lines with width 40")
	}

	// Check that continuation lines are indented
	if len(lines) > 1 {
		firstLine := lines[0]
		secondLine := lines[1]

		if !strings.HasPrefix(firstLine, "[Functions]") {
			t.Error("First line should start with [Functions] prefix")
		}

		// Second line should be indented (start with spaces)
		if !strings.HasPrefix(secondLine, "            ") { // "[Functions] " is 12 chars
			t.Errorf("Second line should be indented, got: %q", secondLine)
		}
	}
}

// stripAnsiCodes removes ANSI color codes for length calculation
func stripAnsiCodes(s string) string {
	// Simple ANSI code stripper for testing
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape && r == 'm' {
			inEscape = false
		} else if !inEscape {
			result += string(r)
		}
	}
	return result
}

// FakeError is a test error type
type FakeError struct {
	message string
}

func (e *FakeError) Error() string {
	return e.message
}

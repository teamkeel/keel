# ViewWriter API Documentation

The `ViewWriter` is a helper for writing text in TUI views with automatic wrapping based on terminal width. This improves developer experience by eliminating the need to pass width on every function call.

## Overview

Before ViewWriter, every render function needed the width passed explicitly:
```go
// Old way - repetitive and error-prone
b.WriteString(wrapText(errorMsg, m.width))
b.WriteString(wrapText(details, m.width))
b.WriteString(wrapText(context, m.width))
```

With ViewWriter, width is set once and all operations respect it:
```go
// New way - clean and ergonomic
w := NewViewWriter(m.width)
w.WriteWrapped(errorMsg)
w.WriteWrapped(details)
w.WriteWrapped(context)
```

## Smart Wrapping

ViewWriter uses **smart wrapping** by default, which intelligently handles both regular text and long paths/URLs:

1. **Word Wrapping**: Breaks text at word boundaries (spaces) when possible
2. **Hard Wrapping**: Automatically breaks long words/paths that exceed terminal width
3. **Hybrid Approach**: Combines both for optimal readability

**Example:**
```go
w := NewViewWriter(50)

// Regular text - wraps at words
w.WriteWrapped("This is a normal sentence that will wrap nicely at word boundaries")
// Output:
// This is a normal sentence that will wrap
// nicely at word boundaries

// Long path - wraps anywhere when no spaces
w.WriteWrapped("/Users/username/Developer/very/long/path/to/node_modules/@teamkeel/functions-runtime/dist/index.cjs")
// Output:
// /Users/username/Developer/very/long/path/to/no
// de_modules/@teamkeel/functions-runtime/dist/i
// ndex.cjs

// Mixed content - best of both
w.WriteWrapped("Error: Cannot find module '/very/long/path/to/module.js' in current directory")
// Output:
// Error: Cannot find module
// '/very/long/path/to/module.js' in current
// directory
```

## API Reference

### Creating a ViewWriter

```go
w := NewViewWriter(width int) *ViewWriter
```

Creates a new ViewWriter with the given terminal width. If width is 0 or negative, defaults to 80.

### Methods

#### `Write(s string)`
Writes a string directly without any wrapping.

```go
w.Write("Fixed width text\n")
w.Write("Another line")
```

**Use when:** You have short text that won't exceed terminal width.

---

#### `Writef(format string, args ...interface{})`
Writes a formatted string without wrapping (like `fmt.Sprintf`).

```go
w.Writef("Processing file %d of %d\n", current, total)
```

**Use when:** You need formatting but text is short.

---

#### `WriteWrapped(text string)`
Writes text with **smart wrapping** - uses word wrapping when possible, but automatically falls back to hard wrapping for long words/paths without spaces.

```go
w.WriteWrapped("This is a very long error message that will automatically wrap to fit within the terminal width and maintain readability.")

// Also handles long paths without spaces
w.WriteWrapped("Error: Cannot find module '/Users/user/very/long/path/to/node_modules/@team/package/index.js'")
```

**Use when:** Text might exceed terminal width (handles both regular text and long paths).

---

#### `WriteWrappedF(format string, args ...interface{})`
Writes formatted text with smart wrapping.

```go
w.WriteWrappedF("Error: Failed to connect to %s on port %d. The server may be down or the connection may be blocked by a firewall.", host, port)

// Also handles formatted paths
w.WriteWrappedF("Cannot find module '%s' in directory '%s'", moduleName, veryLongPath)
```

**Use when:** You need both formatting and wrapping (handles long paths too).

---

#### `WriteWithPrefix(prefix, text string)`
Writes text with a prefix, wrapping and indenting continuation lines to align with the text after the prefix.

```go
w.WriteWithPrefix("[ERROR] ", "A long error message that wraps properly")
// Output:
// [ERROR] A long error message that wraps
//         properly
```

**Use when:** You have prefixed content (logs, errors, list items) that needs proper indentation.

---

#### `Newline()`
Writes a newline character.

```go
w.Newline()
```

**Use when:** You need explicit line breaks.

---

#### `String() string`
Returns the accumulated string.

```go
result := w.String()
```

**Use when:** You're done building the view and need the final string.

## Usage Examples

### Example 1: Simple Error Message

```go
func renderError(m *Model) string {
    w := NewViewWriter(m.width)

    w.Write("❌ Error:\n\n")
    w.WriteWrapped(m.Err.Error())
    w.Newline()

    return w.String()
}
```

### Example 2: Formatted Error with Context

```go
func renderDetailedError(m *Model) string {
    w := NewViewWriter(m.width)

    w.WriteWrappedF("❌ Failed to load schema from %s", m.ProjectDir)
    w.Newline()
    w.Newline()
    w.Write("Details:\n")
    w.WriteWithPrefix("  • ", m.Err.Error())

    return w.String()
}
```

### Example 3: List with Prefixes

```go
func renderIssues(issues []string, width int) string {
    w := NewViewWriter(width)

    w.Write("Found the following issues:\n\n")
    for _, issue := range issues {
        w.WriteWithPrefix(" - ", issue)
        w.Newline()
    }

    return w.String()
}
```

### Example 4: Mixed Content

```go
func renderStatus(m *Model) string {
    w := NewViewWriter(m.width)

    // Short header - no wrapping needed
    w.Writef("Status: %s\n", m.Status)
    w.Newline()

    // Long description - wrap it
    w.WriteWrapped("This is a detailed status message that explains what's happening...")
    w.Newline()
    w.Newline()

    // Logs with prefix and indentation
    for _, log := range m.Logs {
        w.WriteWithPrefix("[LOG] ", log.Message)
        w.Newline()
    }

    return w.String()
}
```

### Example 5: Colored Text with Wrapping

```go
func renderColoredError(err error, width int) string {
    w := NewViewWriter(width)

    // Colors work fine with wrapping
    w.WriteWrapped(colors.Red("Error: ").String() + err.Error())

    return w.String()
}
```

## Before & After Comparison

### Before (Old API)

```go
func renderError(m *Model) string {
    b := strings.Builder{}

    b.WriteString("❌ There was an error:\n\n")
    b.WriteString(wrapText(m.Err.Error(), m.width))
    b.WriteString("\n")

    if m.Details != "" {
        b.WriteString("\nDetails:\n")
        b.WriteString(wrapText(m.Details, m.width))
        b.WriteString("\n")
    }

    for _, suggestion := range m.Suggestions {
        prefix := " - "
        wrapped := wrapText(suggestion, m.width-len(prefix))
        lines := strings.Split(wrapped, "\n")
        b.WriteString(prefix)
        b.WriteString(lines[0])
        if len(lines) > 1 {
            indent := strings.Repeat(" ", len(prefix))
            for i := 1; i < len(lines); i++ {
                b.WriteString("\n")
                b.WriteString(indent)
                b.WriteString(lines[i])
            }
        }
        b.WriteString("\n")
    }

    return b.String()
}
```

### After (New API with ViewWriter)

```go
func renderError(m *Model) string {
    w := NewViewWriter(m.width)

    w.Write("❌ There was an error:\n\n")
    w.WriteWrapped(m.Err.Error())
    w.Newline()

    if m.Details != "" {
        w.Write("\nDetails:\n")
        w.WriteWrapped(m.Details)
        w.Newline()
    }

    for _, suggestion := range m.Suggestions {
        w.WriteWithPrefix(" - ", suggestion)
        w.Newline()
    }

    return w.String()
}
```

**Benefits:**
- ✅ 15 lines → 11 lines (27% reduction)
- ✅ No manual indent calculation
- ✅ No manual line splitting
- ✅ Width only specified once
- ✅ More readable and maintainable

## Best Practices

### 1. Create ViewWriter at function start
```go
func renderSomething(m *Model) string {
    w := NewViewWriter(m.width)  // ✅ Once at the top
    // ... use w throughout
    return w.String()
}
```

### 2. Use appropriate methods for content length

```go
// Short text - use Write()
w.Write("Status: OK\n")

// Long text - use WriteWrapped()
w.WriteWrapped(longErrorMessage)

// Prefixed long text - use WriteWithPrefix()
w.WriteWithPrefix("[ERROR] ", longErrorMessage)
```

### 3. Combine with existing color helpers

```go
// Colors and wrapping work together
w.WriteWrapped(colors.Red("Error: ").String() + errorMsg)
w.WriteWithPrefix(colors.Yellow("[WARN] ").String(), warning)
```

### 4. Use Writef/WriteWrappedF for formatting

```go
// Short formatted text
w.Writef("Count: %d\n", count)

// Long formatted text
w.WriteWrappedF("Error processing file %s: %v", filename, err)
```

## Testing

The ViewWriter has comprehensive test coverage in `views_test.go`:

```bash
# Run ViewWriter tests
go test ./cmd/program -v -run TestViewWriter
```

Tests cover:
- Basic writing
- Formatted writing
- Text wrapping
- Prefix with indentation
- Mixed usage
- Edge cases (zero width, empty strings)

## Migration Guide

To migrate existing code to use ViewWriter:

1. **Replace `strings.Builder` with `ViewWriter`**
   ```go
   // Before
   b := strings.Builder{}

   // After
   w := NewViewWriter(m.width)
   ```

2. **Replace basic writes**
   ```go
   // Before
   b.WriteString("text")

   // After
   w.Write("text")
   ```

3. **Replace wrapped writes**
   ```go
   // Before
   b.WriteString(wrapText(text, m.width))

   // After
   w.WriteWrapped(text)
   ```

4. **Replace prefixed writes**
   ```go
   // Before
   prefix := "[ERROR] "
   wrapped := wrapText(text, m.width-len(prefix))
   lines := strings.Split(wrapped, "\n")
   // ... manual indentation logic ...

   // After
   w.WriteWithPrefix("[ERROR] ", text)
   ```

5. **Replace final return**
   ```go
   // Before
   return b.String()

   // After
   return w.String()
   ```

## Performance

ViewWriter is designed to be efficient:
- Uses `strings.Builder` internally (efficient string concatenation)
- Wrapping calculations are O(n) where n is text length
- Minimal memory allocations
- Suitable for rendering in hot paths

## Summary

The ViewWriter provides:
- **Better DX:** Width specified once, used everywhere
- **Cleaner Code:** Less boilerplate, more readable
- **Automatic Wrapping:** No manual text width management
- **Smart Indentation:** Continuation lines align properly
- **Type Safety:** Methods guide correct usage
- **Well Tested:** 9 comprehensive tests

Use ViewWriter for all new TUI rendering code to maintain consistency and readability!

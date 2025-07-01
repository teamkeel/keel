# Keel Schema Validation Panic Fix

## Issue Description

The keel repository was experiencing a runtime panic with a nil pointer dereference during schema validation:

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x2 addr=0xe0 pc=0x101291744]

goroutine 1 [running]:
github.com/teamkeel/keel/schema/query.ModelFields(0x1400029d2d0?, {0x0, 0x0, 0x1?})
    /home/runner/work/keel/keel/schema/query/query.go:309 +0x24
```

## Root Cause Analysis

The panic occurred in the `ModelFields` function when it tried to iterate over `model.Sections` where `model` was `nil`. The issue originated from the `ComputedNullableFieldRules` validation function:

1. During validation of computed fields, the code iterates through field lookup expressions
2. When processing nested field references, it calls `currModel = query.Model(asts, currField.Type.Value)`
3. If the field type doesn't exist in the schema, `query.Model` returns `nil`
4. On the next iteration, `query.Field(currModel, ident)` is called with `currModel = nil`
5. The `Field` function calls `ModelFields(model)` without checking if the model is nil
6. `ModelFields` tries to access `model.Sections` but crashes because `model` is nil

## Fix Applied

Added nil checks to all functions in `schema/query/query.go` that directly access `model.Sections`:

### 1. `Field` function
```go
func Field(model *parser.ModelNode, name string) *parser.FieldNode {
+   if model == nil {
+       return nil
+   }
    for _, f := range ModelFields(model) {
        // ... existing code
    }
}
```

### 2. `ModelFields` function
```go
func ModelFields(model *parser.ModelNode, filters ...ModelFieldFilter) (res []*parser.FieldNode) {
+   if model == nil {
+       return res
+   }
    for _, section := range model.Sections {
        // ... existing code
    }
}
```

### 3. `ModelField` function
```go
func ModelField(model *parser.ModelNode, name string) *parser.FieldNode {
+   if model == nil {
+       return nil
+   }
    for _, section := range model.Sections {
        // ... existing code
    }
}
```

### 4. `ModelAttributes` function
```go
func ModelAttributes(model *parser.ModelNode) (res []*parser.AttributeNode) {
+   if model == nil {
+       return res
+   }
    for _, section := range model.Sections {
        // ... existing code
    }
}
```

### 5. `ModelActions` function
```go
func ModelActions(model *parser.ModelNode, filters ...ModelActionFilter) (res []*parser.ActionNode) {
+   if model == nil {
+       return res
+   }
    for _, section := range model.Sections {
        // ... existing code
    }
}
```

## Verification

The fix was verified by creating a test that calls all the fixed functions with `nil` parameters. All functions now handle `nil` input gracefully:

- `query.Field(nil, "test")` returns `nil` instead of panicking
- `query.ModelFields(nil)` returns empty slice instead of panicking
- `query.ModelField(nil, "test")` returns `nil` instead of panicking
- `query.ModelAttributes(nil)` returns empty slice instead of panicking
- `query.ModelActions(nil)` returns empty slice instead of panicking

## Impact

This fix resolves the segmentation fault that was preventing the keel generate command from completing successfully. The validation process can now handle cases where field types reference non-existent models without crashing.
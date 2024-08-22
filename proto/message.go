package proto

import "github.com/samber/lo"

// HasFiles checks if the message has any Inline file fields
func (m *Message) HasFiles() bool {
	return len(m.FileFields()) > 0
}

// FileFields will return a slice of fields for the model that are of type file
func (m *Message) FileFields() []*MessageField {
	return lo.Filter(m.Fields, func(f *MessageField, _ int) bool {
		return f.IsFile()
	})
}

// IsModelField returns true if the input targets a model field
// and is handled automatically by the runtime.
// This will only be true for inputs that are built-in actions,
// as functions never have this behaviour.
func (f *MessageField) IsModelField() bool {
	return len(f.Target) > 0
}

// IsFile tells us if the field is a file
func (f *MessageField) IsFile() bool {
	if f.Type == nil {
		return false
	}

	return f.Type.Type == Type_TYPE_INLINE_FILE
}

func (m *Message) FindField(fieldName string) *MessageField {
	for _, field := range m.Fields {
		if field.Name == fieldName {
			return field
		}
	}

	return nil
}

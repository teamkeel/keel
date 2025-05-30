package proto

// IsModelField returns true if the input targets a model field
// and is handled automatically by the runtime.
// This will only be true for inputs that are built-in actions,
// as functions never have this behaviour.
func (f *MessageField) IsModelField() bool {
	return len(f.GetTarget()) > 0
}

// IsFile tells us if the field is a file.
func (f *MessageField) IsFile() bool {
	if f.GetType() == nil {
		return false
	}

	return f.GetType().GetType() == Type_TYPE_FILE
}

// IsMessage checks if the field is a message itself.
func (f *MessageField) IsMessage() bool {
	return f.GetType().GetType() == Type_TYPE_MESSAGE
}

func (m *Message) FindField(fieldName string) *MessageField {
	for _, field := range m.GetFields() {
		if field.GetName() == fieldName {
			return field
		}
	}

	return nil
}

// GetOrderByField returns the orderBy message field, if it has any; otherwise returns nil;.
func (m *Message) GetOrderByField() *MessageField {
	for _, field := range m.GetFields() {
		if field.GetName() == "orderBy" && field.GetType().GetType() == Type_TYPE_UNION {
			return field
		}
	}

	return nil
}

func messageHasFiles(s *Schema, f *Message) bool {
	for _, field := range f.GetFields() {
		if field.GetType().GetMessageName() != nil {
			message := s.FindMessage(field.GetType().GetMessageName().GetValue())
			return messageHasFiles(s, message)
		} else if field.IsFile() {
			return true
		}
	}
	return false
}

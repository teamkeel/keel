package proto

// IsFile tells us if the field is a file
func (f *Field) IsFile() bool {
	if f.Type == nil {
		return false
	}

	return f.Type.Type == Type_TYPE_INLINE_FILE
}

// IsTypeModel returns true of the field's type is Model.
func (f *Field) IsTypeModel() bool {
	return f.Type.Type == Type_TYPE_MODEL
}

// IsTypeRepeated returns true if the field is specified as
// being "repeated".
func (f *Field) IsRepeated() bool {
	return f.Type.Repeated
}

func (f *Field) IsHasMany() bool {
	return f.Type.Type == Type_TYPE_MODEL && f.ForeignKeyFieldName == nil && f.Type.Repeated
}

func (f *Field) IsHasOne() bool {
	return f.Type.Type == Type_TYPE_MODEL && f.ForeignKeyFieldName == nil && !f.Type.Repeated
}

func (f *Field) IsBelongsTo() bool {
	return f.Type.Type == Type_TYPE_MODEL && f.ForeignKeyFieldName != nil && !f.Type.Repeated
}

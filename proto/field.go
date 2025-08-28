package proto

// IsFile tells us if the field is a file.
func (f *Field) IsFile() bool {
	if f.GetType() == nil {
		return false
	}

	return f.GetType().GetType() == Type_TYPE_FILE
}

// IsTypeModel returns true of the field's type is Model.
func (f *Field) IsTypeModel() bool {
	return f.GetType().GetType() == Type_TYPE_ENTITY
}

// IsTypeRepeated returns true if the field is specified as
// being "repeated".
func (f *Field) IsRepeated() bool {
	return f.GetType().GetRepeated()
}

func (f *Field) IsHasMany() bool {
	return f.GetType().GetType() == Type_TYPE_ENTITY && f.GetForeignKeyFieldName() == nil && f.GetType().GetRepeated()
}

func (f *Field) IsHasOne() bool {
	return f.GetType().GetType() == Type_TYPE_ENTITY && f.GetForeignKeyFieldName() == nil && !f.GetType().GetRepeated()
}

func (f *Field) IsBelongsTo() bool {
	return f.GetType().GetType() == Type_TYPE_ENTITY && f.GetForeignKeyFieldName() != nil && !f.GetType().GetRepeated()
}

func (f *Field) IsForeignKey() bool {
	return f.GetForeignKeyInfo() != nil
}

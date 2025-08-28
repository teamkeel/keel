package parser

type Entity interface {
	GetName() string
	Field(name string) *FieldNode
	Fields() []*FieldNode
	GetAttributes() []*AttributeNode
	IsBuiltIn() bool
	EntityType() string
	Node() EntityNode
}

func (m *ModelNode) GetName() string {
	return m.Name.Value
}

func (m *ModelNode) Node() EntityNode {
	return m.EntityNode
}

func (m *ModelNode) Fields() (res []*FieldNode) {
	for _, section := range m.Sections {
		if section.Fields == nil {
			continue
		}
		res = append(res, section.Fields...)
	}

	return res
}

func (m *ModelNode) Field(name string) *FieldNode {
	for _, section := range m.Sections {
		for _, field := range section.Fields {
			if field.Name.Value == name {
				return field
			}
		}
	}
	return nil
}

func (m *ModelNode) GetAttributes() (res []*AttributeNode) {
	for _, section := range m.Sections {
		if section.Attribute != nil {
			res = append(res, section.Attribute)
		}
	}

	return res
}

func (m *ModelNode) IsBuiltIn() bool {
	return m.BuiltIn
}

func (m *ModelNode) EntityType() string {
	return "model"
}

func (t *TaskNode) GetName() string {
	return t.Name.Value
}

func (t *TaskNode) Field(name string) *FieldNode {
	for _, section := range t.Sections {
		for _, field := range section.Fields {
			if field.Name.Value == name {
				return field
			}
		}
	}
	return nil
}

func (t *TaskNode) Fields() (res []*FieldNode) {
	for _, section := range t.Sections {
		if section.Fields == nil {
			continue
		}
		res = append(res, section.Fields...)
	}

	return res
}

func (t *TaskNode) GetAttributes() (res []*AttributeNode) {
	for _, section := range t.Sections {
		if section.Attribute != nil {
			res = append(res, section.Attribute)
		}
	}

	return res
}

func (t *TaskNode) IsBuiltIn() bool {
	return false
}

func (t *TaskNode) EntityType() string {
	return "task"
}

func (t *TaskNode) Node() EntityNode {
	return t.EntityNode
}

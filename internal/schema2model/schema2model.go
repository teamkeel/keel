package schema2model

import (
	"fmt"

	"github.com/teamkeel/keel/parser"
)

type Schema2Model struct{
	keelSchema string
}

func NewSchema2Model(keelSchema string) *Schema2Model {
	return &Schema2Model{
		keelSchema: keelSchema,
	}
}

func (scm *Schema2Model) Parse() (*parser.Schema, error) {
	declarationsAST, err := parser.Parse(scm.keelSchema)
	if err != nil {
		return nil, fmt.Errorf("parser.Parse() failed with: %v", err)
	}
	return declarationsAST, nil
}
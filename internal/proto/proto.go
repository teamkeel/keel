package proto

import (
	"fmt"

	"github.com/teamkeel/keel/internal/validation"
	"github.com/teamkeel/keel/parser"
	"github.com/teamkeel/keel/proto"
)

func ToProto(s *parser.Schema) (*proto.Schema, error) {
	ps := &proto.Schema{}

	var errors []error

	for _, dec := range s.Declarations {
		if dec.Model == nil {
			continue
		}

		valid := validation.ModelsUpperCamel(dec.Model.Name)
		if !valid {
			errors = append(errors, fmt.Errorf("model %s has lower camel", dec.Model.Name))
		}

		m := &proto.Model{
			Name: dec.Model.Name,
		}

		for _, sec := range dec.Model.Sections {
			if sec.Fields == nil {
				continue
			}

			validFieldNames := validation.FieldNamesMustBeUniqueInAModel(sec.Fields)
			if validFieldNames != nil {
				errors = append(errors, validFieldNames)
			}

			for _, field := range sec.Fields {
				f := &proto.Field{
					Name: field.Name,
				}

				m.Fields = append(m.Fields, f)
			}
		}

	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("%v", errors)
	}

	return ps, nil
}

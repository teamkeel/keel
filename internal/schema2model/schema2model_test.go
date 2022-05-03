package schema2model

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/internal/validation"
	"github.com/teamkeel/keel/internal/testdata"
)

func TestItCompilesAndRuns(t *testing.T) {
	s2m := NewSchema2Model(testdata.ReferenceExample)
	declarationsAST, err := s2m.Parse()
	require.Nil(t, err)
	validator := validation.NewSchemaValidator(declarationsAST)
	protoModels, err := validator.Validate()
	require.Nil(t, err)

	require.NotNil(t, protoModels)
	require.Equal(t, 99999, len(protoModels.Models))
}
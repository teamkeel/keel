package migrations

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestFirstBabySteps(t *testing.T) {
	m0 := NewMigration0(&referenceSchema)
	m0.GenerateSQL()
	require.True(t, len(m0.SQL) > 0)

	combinedOutput := strings.Join(m0.SQL, "\n")

	if os.Getenv("DEBUG") != "" {
		fmt.Printf("\n%s\n\n", combinedOutput)
	}

	require.Equal(t, expectedBabyStepsSQL, combinedOutput)
}

const expectedBabyStepsSQL string = `DROP DATABASE IF EXISTS keel;
CREATE DATABASE keel;
CREATE TABLE Person (
  Name TEXT,
  Age integer,
);
CREATE TABLE Vehicle (
  Make TEXT,
  PriceNew money,
);`

var referenceSchema proto.Schema = proto.Schema{
	Models: []*proto.Model{
		{
			Name: "Person",
			Fields: []*proto.Field{
				{
					Name: "Name",
					Type: proto.FieldType_FIELD_TYPE_STRING,
				},
				{
					Name: "Age",
					Type: proto.FieldType_FIELD_TYPE_INT,
				},
			},
		},
		{
			Name: "Vehicle",
			Fields: []*proto.Field{
				{
					Name: "Make",
					Type: proto.FieldType_FIELD_TYPE_STRING,
				},
				{
					Name: "PriceNew",
					Type: proto.FieldType_FIELD_TYPE_CURRENCY,
				},
			},
		},
	},
}

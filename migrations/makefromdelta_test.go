package migrations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestItFindsTheDifferencesItShould(t *testing.T) {
	generatedSQL, err := MakeMigrationsFromSchemaDifference(&oldProto, &newProto)
	require.NoError(t, err)
	if generatedSQL != expected {
		fmt.Printf("\n\n%s\n\n", generatedSQL)
	}
	require.Equal(t, expected, generatedSQL)
}

var oldProto proto.Schema = proto.Schema{
	Models: []*proto.Model{
		{
			Name: "Person",
			Fields: []*proto.Field{
				{
					Name: "Name",
					Type: &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
				},
				{
					Name: "Age",
					Type: &proto.TypeInfo{Type: proto.Type_TYPE_INT},
				},
			},
		},

		{
			Name: "Address",
			Fields: []*proto.Field{
				{
					Name: "Postcode",
					Type: &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
				},
			},
		},
	},
}

// New Proto is a copy of oldProto - to which the following changes have been applied:
//
// o  The <Person> model has been renamed to <Human> // Drop one table, create another.
// o  The field Address.Postcode has been renamed to <City>. // Drop one field, create another.
var newProto proto.Schema = proto.Schema{
	Models: []*proto.Model{
		{
			Name: "Human",
			Fields: []*proto.Field{
				{
					Name: "Name",
					Type: &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
				},
				{
					Name: "Age",
					Type: &proto.TypeInfo{Type: proto.Type_TYPE_INT},
				},
			},
		},

		{
			Name: "Address",
			Fields: []*proto.Field{
				{
					Name: "City",
					Type: &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
				},
			},
		},
	},
}

const expected string = `
CREATE TABLE "Human"(
"Name" TEXT,
"Age" INTEGER
);
DROP TABLE "Person";
ALTER TABLE "Address" ADD COLUMN "City" TEXT;
ALTER TABLE "Address" DROP COLUMN "Postcode";`

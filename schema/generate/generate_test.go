package generate_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema/generate"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/reader"
)

func TestGenerateDefaultActionsAsStrings(t *testing.T) {
	tests := []struct {
		name            string
		schema          string
		modelName       string
		expectedActions []string
	}{
		{
			name: "Patient model with various field types",
			schema: `
enum SkinSize {
    Small
    Medium
    Large
}

model PatientPrescriber {
    fields {
        name Text
    }
}

model Patient {
    fields {
        firstName Text
        lastName Text
        dob Date
        patientNumber ID @unique
        pregnant Boolean?
        allergies Text?
        address1 Text?
        address2 Text?
        city Text?
        county Text?
        country Text?
        mobile Text?
        skinSize SkinSize?
        prescribers PatientPrescriber[]
    }
}`,
			modelName: "Patient",
			expectedActions: []string{
				"get getPatient(id)",
				`list listPatients(
    firstName?,
    lastName?,
    dob?,
    patientNumber?,
    pregnant?,
    allergies?,
    address1?,
    address2?,
    city?,
    county?,
    country?,
    mobile?,
    skinSize?,
    prescribers.id?
)`,
				`create createPatient() with (
    firstName,
    lastName,
    dob,
    patientNumber,
    pregnant?,
    allergies?,
    address1?,
    address2?,
    city?,
    county?,
    country?,
    mobile?,
    skinSize?
)`,
				`update updatePatient(id) with (
    firstName?,
    lastName?,
    dob?,
    patientNumber?,
    pregnant?,
    allergies?,
    address1?,
    address2?,
    city?,
    county?,
    country?,
    mobile?,
    skinSize?
)`,
				"delete deletePatient(id)",
			},
		},
		{
			name: "Simple model with only required fields",
			schema: `
model User {
    fields {
        name Text
        email Text
    }
}`,
			modelName: "User",
			expectedActions: []string{
				"get getUser(id)",
				"list listUsers(name?, email?)",
				"create createUser() with (name, email)",
				"update updateUser(id) with (name?, email?)",
				"delete deleteUser(id)",
			},
		},
		{
			name: "Model with only optional fields",
			schema: `
model Profile {
    fields {
        bio Text?
        avatar Text?
    }
}`,
			modelName: "Profile",
			expectedActions: []string{
				"get getProfile(id)",
				"list listProfiles(bio?, avatar?)",
				"create createProfile() with (bio?, avatar?)",
				"update updateProfile(id) with (bio?, avatar?)",
				"delete deleteProfile(id)",
			},
		},
		{
			name: "Model with relationships",
			schema: `
model Author {
    fields {
        name Text
    }
}

model Book {
    fields {
        title Text
        author Author
        coAuthors Author[]
    }
}`,
			modelName: "Book",
			expectedActions: []string{
				"get getBook(id)",
				"list listBooks(title?, author.id?, coAuthors.id?)",
				"create createBook() with (title, author.id)",
				"update updateBook(id) with (title?, author.id?)",
				"delete deleteBook(id)",
			},
		},
		{
			name: "Model with no fields",
			schema: `
model Empty {
    fields {
    }
}`,
			modelName: "Empty",
			expectedActions: []string{
				"get getEmpty(id)",
				"list listEmpties()",
				"create createEmpty()",
				"update updateEmpty(id)",
				"delete deleteEmpty(id)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the schema
			ast, err := parser.Parse(&reader.SchemaFile{
				FileName: "test.keel",
				Contents: tt.schema,
			})
			require.NoError(t, err)

			// Generate actions as strings
			actualActions := generate.GenerateDefaultActionsAsStrings([]*parser.AST{ast}, tt.modelName)
			require.Len(t, actualActions, len(tt.expectedActions))

			for i, expected := range tt.expectedActions {
				assert.Equal(t, normalizeWhitespace(expected), normalizeWhitespace(actualActions[i]),
					"Action %d mismatch\nExpected: %s\nActual: %s", i, expected, actualActions[i])
			}
		})
	}
}

func TestGenerateDefaultActionsNonExistentModel(t *testing.T) {
	schema := `
model User {
    fields {
        name Text
    }
}`

	ast, err := parser.Parse(&reader.SchemaFile{
		FileName: "test.keel",
		Contents: schema,
	})
	require.NoError(t, err)

	// Test with non-existent model
	actions := generate.GenerateDefaultActions([]*parser.AST{ast}, "NonExistent")
	assert.Empty(t, actions)
}

// normalizeWhitespace normalizes whitespace for easier comparison
func normalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var normalizedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalizedLines = append(normalizedLines, trimmed)
		}
	}
	return strings.Join(normalizedLines, "\n")
}

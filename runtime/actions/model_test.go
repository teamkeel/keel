package actions

import (
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestBuiltInDefaultEqualities(t *testing.T) {
	for _, tt := range equalities {
		repeated := false
		v, err := builtinDefault(field(tt.protoType, repeated), someEnums)
		require.NoError(t, err)
		if tt.expectedHasOne != v {
			t.Fatalf("For type %s, expected %v, got %v, fixture is: ", tt.protoType, tt.expectedHasOne, v)
		}

		repeated = true
		v, err = builtinDefault(field(tt.protoType, repeated), someEnums)
		require.NoError(t, err)
		require.EqualValues(t, tt.expectedHasMany, v, "case is: %+v", tt)
	}
}

func TestBuiltInDefaultForID(t *testing.T) {
	repeated := false
	v, err := builtinDefault(field(proto.Type_TYPE_ID, repeated), someEnums)
	require.NoError(t, err)
	// Make sure it is a KSUID
	id, ok := v.(ksuid.KSUID)
	require.True(t, ok)
	// Make sure it encodes a very recent time.
	timeSinceMade := time.Since(id.Time())
	require.Less(t, timeSinceMade, 5*time.Second)

	repeated = true
	v, err = builtinDefault(field(proto.Type_TYPE_ID, repeated), someEnums)
	require.NoError(t, err)
	ids, ok := v.([]ksuid.KSUID)
	require.True(t, ok)
	require.Len(t, ids, 0)
}

func TestBuiltInDefaultForTimeFields(t *testing.T) {
	toTest := []proto.Type{
		proto.Type_TYPE_DATE,
		proto.Type_TYPE_DATETIME,
		proto.Type_TYPE_TIMESTAMP,
	}
	for _, fieldType := range toTest {
		repeated := false
		v, err := builtinDefault(field(fieldType, repeated), someEnums)
		require.NoError(t, err)
		timeEncoded, ok := v.(time.Time)
		require.True(t, ok)
		timeSinceMade := time.Since(timeEncoded)
		require.Less(t, timeSinceMade, 5*time.Second)

		repeated = true
		v, err = builtinDefault(field(fieldType, repeated), someEnums)
		require.NoError(t, err)
		times, ok := v.([]time.Time)
		require.True(t, ok)
		require.Len(t, times, 9)
	}
}

func TestBuiltInDefaultForEnumIsNil(t *testing.T) {
	repeated := false
	enumField := field(proto.Type_TYPE_ENUM, repeated)
	enumField.Type.EnumName = wrapperspb.String("fruits")
	v, err := builtinDefault(enumField, someEnums)
	require.NoError(t, err)
	require.Nil(t, v)

	repeated = true
	enumField = field(proto.Type_TYPE_ENUM, repeated)
	enumField.Type.EnumName = wrapperspb.String("fruits")
	v, err = builtinDefault(enumField, someEnums)
	require.NoError(t, err)
	require.Nil(t, v)
}

type defaultValueCase struct {
	name         string
	protoType    proto.Type
	repeated     bool
	defaultValue string
	expected     any
}

func TestSchemaDefaults(t *testing.T) {

	const aTimestamp string = "2006-01-02T15:04:05Z"
	const layout string = time.RFC3339
	stampAsTime, err := time.Parse(layout, aTimestamp)
	require.NoError(t, err)

	cases := []defaultValueCase{
		{
			name:         "string",
			protoType:    proto.Type_TYPE_STRING,
			defaultValue: `"my default string"`,
			expected:     `my default string`,
		},
		{
			name:         "false",
			protoType:    proto.Type_TYPE_BOOL,
			defaultValue: `false`,
			expected:     false,
		},
		{
			name:         "true",
			protoType:    proto.Type_TYPE_BOOL,
			defaultValue: `true`,
			expected:     true,
		},
		{
			name:         "int",
			protoType:    proto.Type_TYPE_INT,
			defaultValue: `42`,
			expected:     int64(42),
		},
		{
			name:         "date",
			protoType:    proto.Type_TYPE_DATE,
			defaultValue: `"30/06/2011"`,
			expected:     time.Date(2011, time.June, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "datetime",
			protoType:    proto.Type_TYPE_DATETIME,
			defaultValue: `"` + aTimestamp + `"`,
			expected:     stampAsTime,
		},
		{
			name:         "timestamp",
			protoType:    proto.Type_TYPE_TIMESTAMP,
			defaultValue: `"` + aTimestamp + `"`,
			expected:     stampAsTime,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			f := field(cs.protoType, cs.repeated)
			f.DefaultValue = &proto.DefaultValue{
				Expression: &proto.Expression{
					Source: cs.defaultValue,
				},
			}
			v, err := schemaDefault(f)
			require.NoError(t, err)
			require.Equal(t, cs.expected, v)
		})
	}
}

func TestErrorNotYetSupported(t *testing.T) {
	repeated := false
	f := field(proto.Type_TYPE_BOOL, repeated)
	f.DefaultValue = &proto.DefaultValue{
		Expression: &proto.Expression{
			Source: "True == False", // We haven't implemented the evaluation of expressions that are not simple values.
		},
	}
	_, err := schemaDefault(f)
	require.EqualError(t, err, "expressions that are not simple values are not yet supported")

}

func TestInitialValueForFieldPrefersSchemaDefault(t *testing.T) {
	// set up field with both default expr and use-zero
	repeated := false
	f := field(proto.Type_TYPE_STRING, repeated)
	f.DefaultValue = &proto.DefaultValue{
		Expression: &proto.Expression{
			Source: `"hello expression"`,
		},
		UseZeroValue: true,
	}
	// Make sure it uses the schema-default
	v, err := initialValueForField(f, someEnums)
	require.NoError(t, err)
	require.Equal(t, "hello expression", v)
}

func TestInitialValueForFieldUsesZeroValueWhenNoSchemaDefault(t *testing.T) {
	// set up field with only use-zero
	repeated := false
	f := field(proto.Type_TYPE_STRING, repeated)
	f.DefaultValue = &proto.DefaultValue{
		Expression:   nil,
		UseZeroValue: true,
	}
	v, err := initialValueForField(f, someEnums)
	require.NoError(t, err)
	require.Equal(t, "", v)
}

func TestInitialValueForFieldUsesNilWhenNoDefaultIsAvailable(t *testing.T) {
	repeated := false
	f := field(proto.Type_TYPE_MODEL, repeated)
	f.DefaultValue = nil
	v, err := initialValueForField(f, someEnums)
	require.NoError(t, err)
	require.Nil(t, v)
}

func TestZeroValueForModel(t *testing.T) {
	// This test should not duplicate the work of the tests that check
	// the zero value for each field type. But should prove that the
	// function under test has assembled the map[string]any that it returns
	// with fields correctly initialised to their zero values. So we check just
	// a sample of two fields with different types.
	model := proto.Model{
		Fields: []*proto.Field{
			{
				Type: &proto.TypeInfo{Type: proto.Type_TYPE_INT},
				Name: "intField",
				DefaultValue: &proto.DefaultValue{
					UseZeroValue: true,
				},
			},
			{
				Type: &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
				Name: "stringField",
				DefaultValue: &proto.DefaultValue{
					UseZeroValue: true,
				},
			},
		},
	}
	var schema *proto.Schema = &proto.Schema{
		Enums: someEnums,
	}
	modelMap, err := initialValueForModel(&model, schema)
	require.NoError(t, err)
	require.Equal(t, 0, modelMap["int_field"])
	require.Equal(t, "", modelMap["string_field"])
}

type equality struct {
	protoType       proto.Type
	expectedHasOne  any
	expectedHasMany any
}

var equalities []equality = []equality{
	{
		protoType:       proto.Type_TYPE_STRING,
		expectedHasOne:  "",
		expectedHasMany: []string{},
	},
	{
		protoType:       proto.Type_TYPE_BOOL,
		expectedHasOne:  false,
		expectedHasMany: []bool{},
	},
	{
		protoType:       proto.Type_TYPE_INT,
		expectedHasOne:  0,
		expectedHasMany: []int{},
	},
	{
		protoType:       proto.Type_TYPE_MODEL,
		expectedHasOne:  "",
		expectedHasMany: []string{},
	},
	{
		protoType:       proto.Type_TYPE_CURRENCY,
		expectedHasOne:  "",
		expectedHasMany: []string{},
	},
}

func field(fieldType proto.Type, repeated bool) *proto.Field {
	return &proto.Field{
		Type: &proto.TypeInfo{
			Type:     fieldType,
			Repeated: repeated,
		},
		DefaultValue: &proto.DefaultValue{
			UseZeroValue: true,
		},
	}
}

var someEnums []*proto.Enum = []*proto.Enum{
	{
		Name: "fruits",
		Values: []*proto.EnumValue{
			{Name: "apple"},
			{Name: "banana"},
		},
	},
	{
		Name: "tshirts",
		Values: []*proto.EnumValue{
			{Name: "small"},
			{Name: "large"},
		},
	},
}

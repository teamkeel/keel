package actions

import (
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestZeroFieldValuesSimpleEquality(t *testing.T) {
	for _, tt := range equalityTable {
		repeated := false
		v, err := zeroValueForField(field(tt.protoType, repeated), someEnums)
		require.NoError(t, err)
		if tt.expectedHasOne != v {
			t.Fatalf("For type %s, expected %v, got %v", tt.protoType, tt.expectedHasOne, v)
		}

		repeated = true
		v, err = zeroValueForField(field(tt.protoType, repeated), someEnums)
		require.NoError(t, err)
		require.EqualValues(t, tt.expectedHasMany, v)
	}
}

func TestZeroFieldValueForID(t *testing.T) {
	repeated := false
	v, err := zeroValueForField(field(proto.Type_TYPE_ID, repeated), someEnums)
	require.NoError(t, err)
	// Make sure it is a KSUID
	id, ok := v.(ksuid.KSUID)
	require.True(t, ok)
	// Make sure it encodes a very recent time.
	timeSinceMade := time.Since(id.Time())
	require.Less(t, timeSinceMade, 5*time.Second)

	repeated = true
	v, err = zeroValueForField(field(proto.Type_TYPE_ID, repeated), someEnums)
	require.NoError(t, err)
	ids, ok := v.([]ksuid.KSUID)
	require.True(t, ok)
	require.Len(t, ids, 0)
}

func TestZeroFieldValuesTimeBasedFields(t *testing.T) {
	toTest := []proto.Type{
		proto.Type_TYPE_DATE,
		proto.Type_TYPE_DATETIME,
		proto.Type_TYPE_TIMESTAMP,
	}
	for _, fieldType := range toTest {
		repeated := false
		v, err := zeroValueForField(field(fieldType, repeated), someEnums)
		require.NoError(t, err)
		timeEncoded, ok := v.(time.Time)
		require.True(t, ok)
		timeSinceMade := time.Since(timeEncoded)
		require.Less(t, timeSinceMade, 5*time.Second)

		repeated = true
		v, err = zeroValueForField(field(fieldType, repeated), someEnums)
		require.NoError(t, err)
		times, ok := v.([]time.Time)
		require.True(t, ok)
		require.Len(t, times, 0)
	}
}

func TestZeroFieldValuesEnum(t *testing.T) {
	repeated := false
	enumField := field(proto.Type_TYPE_ENUM, repeated)
	enumField.Type.EnumName = wrapperspb.String("fruits")
	v, err := zeroValueForField(enumField, someEnums)
	require.NoError(t, err)
	require.Equal(t, "apple", v)

	repeated = true
	enumField = field(proto.Type_TYPE_ENUM, repeated)
	enumField.Type.EnumName = wrapperspb.String("fruits")
	v, err = zeroValueForField(enumField, someEnums)
	require.NoError(t, err)
	enumValues, ok := v.([]string)
	require.True(t, ok)
	require.Len(t, enumValues, 0)
}

func TestZeroFieldValueForTypeNotImplemented(t *testing.T) {
	repeated := false
	_, err := zeroValueForField(field(proto.Type_TYPE_IDENTITY, repeated), someEnums)
	require.EqualError(t, err, "zero value for field: TYPE_IDENTITY not yet implemented")
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
			},
			{
				Type: &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
				Name: "stringField",
			},
		},
	}
	var schema *proto.Schema = &proto.Schema{
		Enums: someEnums,
	}
	modelMap, err := zeroValueForModel(&model, schema)
	require.NoError(t, err)
	require.Equal(t, 0, modelMap["intField"])
	require.Equal(t, "", modelMap["stringField"])
}

type equality struct {
	protoType       proto.Type
	expectedHasOne  any
	expectedHasMany any
}

var equalityTable []equality = []equality{
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

package actions

import (
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/proto"
)

func TestZeroFieldValuesSimpleEquality(t *testing.T) {
	for _, tt := range equalityTable {
		v, err := zeroValueForField(tt.protoType)
		require.NoError(t, err)
		if tt.expected != v {
			t.Fatalf("For type %s, expected %v, got %v", tt.protoType, tt.expected, v)
		}
	}
}

func TestZeroFieldValueForID(t *testing.T) {
	v, err := zeroValueForField(proto.Type_TYPE_ID)
	require.NoError(t, err)
	// Make sure it is a KSUID
	ksuid, ok := v.(ksuid.KSUID)
	require.True(t, ok)
	// Make sure it encodes a very recent time.
	timeSinceMade := time.Since(ksuid.Time())
	require.Less(t, timeSinceMade, 5*time.Second)
}

func TestZeroFieldValueForTypeNotImplemented(t *testing.T) {
	_, err := zeroValueForField(proto.Type_TYPE_IDENTITY)
	require.EqualError(t, err, "zero value for field type: TYPE_IDENTITY not yet implemented")
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
	modelMap, err := zeroValueForModel(&model)
	require.NoError(t, err)
	require.Equal(t, 0, modelMap["intField"])
	require.Equal(t, "", modelMap["stringField"])
}

type equality struct {
	protoType proto.Type
	expected  any
}

var equalityTable []equality = []equality{
	{
		protoType: proto.Type_TYPE_STRING,
		expected:  "",
	},
	{
		protoType: proto.Type_TYPE_BOOL,
		expected:  false,
	},
	{
		protoType: proto.Type_TYPE_INT,
		expected:  0,
	},
	{
		protoType: proto.Type_TYPE_MODEL,
		expected:  "",
	},
	{
		protoType: proto.Type_TYPE_CURRENCY,
		expected:  "",
	},
	{
		protoType: proto.Type_TYPE_ENUM,
		expected:  "",
	},
	{
		protoType: proto.Type_TYPE_DATE,
		expected:  time.Time{},
	},
	{
		protoType: proto.Type_TYPE_DATETIME,
		expected:  time.Time{},
	},
	{
		protoType: proto.Type_TYPE_TIMESTAMP,
		expected:  time.Time{},
	},
}

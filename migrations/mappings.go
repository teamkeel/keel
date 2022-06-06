package migrations

import "github.com/teamkeel/keel/proto"

var PostgresFieldTypes map[proto.FieldType]string = map[proto.FieldType]string{
	proto.FieldType_FIELD_TYPE_UNKNOWN:      "unknown",
	proto.FieldType_FIELD_TYPE_STRING:       "TEXT",
	proto.FieldType_FIELD_TYPE_BOOL:         "bool",
	proto.FieldType_FIELD_TYPE_INT:          "integer",
	proto.FieldType_FIELD_TYPE_TIMESTAMP:    "TIMESTAMP",
	proto.FieldType_FIELD_TYPE_DATE:         "DATE",
	proto.FieldType_FIELD_TYPE_ID:           "UUID",
	proto.FieldType_FIELD_TYPE_RELATIONSHIP: "UUID", // id of the target
	proto.FieldType_FIELD_TYPE_CURRENCY:     "money",
	proto.FieldType_FIELD_TYPE_DATETIME:     "TIMESTAMP",
	proto.FieldType_FIELD_TYPE_ENUM:         "enum-not-implemented-yet",
	proto.FieldType_FIELD_TYPE_IDENTITY:     "identity-not-implemented-yet",
	proto.FieldType_FIELD_TYPE_IMAGE:        "blob",
}

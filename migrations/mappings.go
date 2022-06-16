package migrations

import "github.com/teamkeel/keel/proto"

var PostgresFieldTypes map[proto.FieldType]string = map[proto.FieldType]string{
	proto.FieldType_FIELD_TYPE_UNKNOWN:      "unknown",
	proto.FieldType_FIELD_TYPE_STRING:       "TEXT",
	proto.FieldType_FIELD_TYPE_BOOL:         "bool",
	proto.FieldType_FIELD_TYPE_INT:          "integer",
	proto.FieldType_FIELD_TYPE_TIMESTAMP:    "TIMESTAMP",
	proto.FieldType_FIELD_TYPE_DATE:         "DATE",
	proto.FieldType_FIELD_TYPE_ID:           "TEXT",
	proto.FieldType_FIELD_TYPE_RELATIONSHIP: "TEXT", // id of the target
	proto.FieldType_FIELD_TYPE_CURRENCY:     "money",
	proto.FieldType_FIELD_TYPE_DATETIME:     "TIMESTAMP",
	proto.FieldType_FIELD_TYPE_ENUM:         "TEXT",
	proto.FieldType_FIELD_TYPE_IDENTITY:     "TEXT", // a relationship to an Identity row
	proto.FieldType_FIELD_TYPE_IMAGE:        "blob",
}

package migrations

import "github.com/teamkeel/keel/proto"

var PostgresFieldTypes map[proto.Type]string = map[proto.Type]string{
	proto.Type_TYPE_UNKNOWN:   "unknown",
	proto.Type_TYPE_STRING:    "TEXT",
	proto.Type_TYPE_BOOL:      "bool",
	proto.Type_TYPE_INT:       "integer",
	proto.Type_TYPE_TIMESTAMP: "TIMESTAMP",
	proto.Type_TYPE_DATE:      "DATE",
	proto.Type_TYPE_ID:        "TEXT",
	proto.Type_TYPE_MODEL:     "TEXT", // id of the target
	proto.Type_TYPE_CURRENCY:  "money",
	proto.Type_TYPE_DATETIME:  "TIMESTAMP",
	proto.Type_TYPE_ENUM:      "TEXT",
	proto.Type_TYPE_IDENTITY:  "TEXT", // a relationship to an Identity row
	proto.Type_TYPE_IMAGE:     "blob",
}

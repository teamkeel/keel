package migrations

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// pushAuditModel adds a (hard coded) KeelAudit model to the given schema.
func pushAuditModel(schema *proto.Schema) {
	schema.Models = append(schema.Models, auditModel())
}

// popAuditModel removes any instances present of a KeelAudit model from the given
// schema.
func popAuditModel(schema *proto.Schema) {
	schema.Models = lo.Reject(schema.Models, func(m *proto.Model, _ int) bool {
		return m.Name == auditModelName
	})
}

// auditModel provides a hard-coded Model to represent the audit table.
//
// Migrations will fire as we would wish, if you edit the fields in this model definition.
func auditModel() *proto.Model {
	mdl := proto.Model{
		Name: auditModelName,
		Fields: []*proto.Field{

			{
				ModelName:  auditModelName,
				Name:       "id",
				PrimaryKey: true,
				DefaultValue: &proto.DefaultValue{
					UseZeroValue: true,
				},
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_ID,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String("id"),
				},
				Optional: false,
			},

			{
				ModelName: auditModelName,
				Name:      "tableName",
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String("tableName"),
				},
				Optional: false,
			},

			{
				ModelName: auditModelName,
				Name:      "op",
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String("op"),
				},
				Optional: false,
			},

			{
				ModelName: auditModelName,
				Name:      auditTableDataField,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String(auditTableDataField),
				},
				Optional: false,
			},

			{
				ModelName: auditModelName,
				Name:      "createdAt",
				DefaultValue: &proto.DefaultValue{
					UseZeroValue: true,
				},
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_TIMESTAMP,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String("createdAt"),
				},
				Optional: false,
			},

			{
				ModelName: auditModelName,
				Name:      "identityId",
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_ID,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String("identityId"),
				},
				Optional: true,
			},

			{
				ModelName: auditModelName,
				Name:      "traceId",
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String("traceId"),
				},
				Optional: true,
			},

			{
				ModelName: auditModelName,
				Name:      "eventProcessedAt",
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_TIMESTAMP,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String("eventProcessedAt"),
				},
				Optional: true,
			},
		},
	}
	return &mdl
}

const auditModelName = "KeelAudit"
const auditTableDataField = "data"

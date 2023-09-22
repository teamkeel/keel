package migrations

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/auditing"
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
				Name:       auditing.ColumnId,
				PrimaryKey: true,
				DefaultValue: &proto.DefaultValue{
					UseZeroValue: true,
				},
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_ID,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String(auditing.ColumnId),
				},
				Optional: false,
			},
			{
				ModelName: auditModelName,
				Name:      auditing.ColumnTableName,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String(auditing.ColumnTableName),
				},
				Optional: false,
			},
			{
				ModelName: auditModelName,
				Name:      auditing.ColumnOp,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String(auditing.ColumnOp),
				},
				Optional: false,
			},
			{
				ModelName: auditModelName,
				Name:      auditing.ColumnData,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String(auditing.ColumnData),
				},
				Optional: false,
			},
			{
				ModelName: auditModelName,
				Name:      auditing.ColumnCreatedAt,
				DefaultValue: &proto.DefaultValue{
					UseZeroValue: true,
				},
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_TIMESTAMP,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String(auditing.ColumnCreatedAt),
				},
				Optional: false,
			},
			{
				ModelName: auditModelName,
				Name:      auditing.ColumnIdentityId,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_ID,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String(auditing.ColumnIdentityId),
				},
				Optional: true,
			},
			{
				ModelName: auditModelName,
				Name:      auditing.ColumnTraceId,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String(auditing.ColumnTraceId),
				},
				Optional: true,
			},
			{
				ModelName: auditModelName,
				Name:      auditing.ColumnEventProcessedAt,
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_TIMESTAMP,
					ModelName: wrapperspb.String(auditModelName),
					FieldName: wrapperspb.String(auditing.ColumnEventProcessedAt),
				},
				Optional: true,
			},
		},
	}
	return &mdl
}

const auditModelName = auditing.TableName

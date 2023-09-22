package migrations

import (
	"github.com/iancoleman/strcase"
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
		return m.Name == strcase.ToCamel(auditing.TableName)
	})
}

// auditModel provides a hard-coded Model to represent the audit table.
//
// Migrations will fire as we would wish, if you edit the fields in this model definition.
func auditModel() *proto.Model {
	modelName := strcase.ToCamel(auditing.TableName)

	mdl := proto.Model{
		Name: modelName,
		Fields: []*proto.Field{
			{
				ModelName:  modelName,
				Name:       strcase.ToLowerCamel(auditing.ColumnId),
				PrimaryKey: true,
				DefaultValue: &proto.DefaultValue{
					UseZeroValue: true,
				},
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_ID,
					ModelName: wrapperspb.String(modelName),
					FieldName: wrapperspb.String(strcase.ToLowerCamel(auditing.ColumnId)),
				},
				Optional: false,
			},
			{
				ModelName: modelName,
				Name:      strcase.ToLowerCamel(auditing.ColumnTableName),
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(modelName),
					FieldName: wrapperspb.String(strcase.ToLowerCamel(auditing.ColumnTableName)),
				},
				Optional: false,
			},
			{
				ModelName: modelName,
				Name:      strcase.ToLowerCamel(auditing.ColumnOp),
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(modelName),
					FieldName: wrapperspb.String(strcase.ToLowerCamel(auditing.ColumnOp)),
				},
				Optional: false,
			},
			{
				ModelName: modelName,
				Name:      strcase.ToLowerCamel(auditing.ColumnData),
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(modelName),
					FieldName: wrapperspb.String(strcase.ToLowerCamel(auditing.ColumnData)),
				},
				Optional: false,
			},
			{
				ModelName: modelName,
				Name:      strcase.ToLowerCamel(auditing.ColumnCreatedAt),
				DefaultValue: &proto.DefaultValue{
					UseZeroValue: true,
				},
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_TIMESTAMP,
					ModelName: wrapperspb.String(modelName),
					FieldName: wrapperspb.String(strcase.ToLowerCamel(auditing.ColumnCreatedAt)),
				},
				Optional: false,
			},
			{
				ModelName: modelName,
				Name:      strcase.ToLowerCamel(auditing.ColumnIdentityId),
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_ID,
					ModelName: wrapperspb.String(modelName),
					FieldName: wrapperspb.String(strcase.ToLowerCamel(auditing.ColumnIdentityId)),
				},
				Optional: true,
			},
			{
				ModelName: modelName,
				Name:      strcase.ToLowerCamel(auditing.ColumnTraceId),
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_STRING,
					ModelName: wrapperspb.String(modelName),
					FieldName: wrapperspb.String(strcase.ToLowerCamel(auditing.ColumnTraceId)),
				},
				Optional: true,
			},
			{
				ModelName: modelName,
				Name:      strcase.ToLowerCamel(auditing.ColumnEventProcessedAt),
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_TIMESTAMP,
					ModelName: wrapperspb.String(modelName),
					FieldName: wrapperspb.String(strcase.ToLowerCamel(auditing.ColumnEventProcessedAt)),
				},
				Optional: true,
			},
		},
	}
	return &mdl
}

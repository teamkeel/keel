package schema

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/cron"
	"github.com/teamkeel/keel/expressions/resolve"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// makeProtoModels derives and returns a proto.Schema from the given (known to be valid) set of parsed AST.
func (scm *Builder) makeProtoModels() *proto.Schema {
	scm.proto = &proto.Schema{}

	// makeAnyType adds a global 'Any' type to the messages registry which is useful for those who want untyped inputs and responses for arbitrary functions
	scm.makeAnyType()

	// Add any messages defined declaratively in the schema to the registry of message types
	for _, ast := range scm.asts {
		for _, d := range ast.Declarations {
			if d.Message != nil {
				scm.makeMessage(d)
			}
		}
	}

	for _, parserSchema := range scm.asts {
		for _, decl := range parserSchema.Declarations {
			switch {
			case decl.Model != nil:
				scm.makeModel(decl)
			case decl.Role != nil:
				scm.makeRole(decl)
			case decl.API != nil:
				scm.makeAPI(decl)
			case decl.Enum != nil:
				scm.makeEnum(decl)
			case decl.Job != nil:
				scm.makeJob(decl)
			case decl.Flow != nil:
				scm.makeFlow(decl)
			case decl.Message != nil:
				// noop
			case decl.Routes != nil:
				scm.makeRoutes(decl)
			default:
				panic("Case not recognized")
			}
		}
	}

	if scm.Config != nil {
		for _, envVar := range scm.Config.AllEnvironmentVariables() {
			scm.proto.EnvironmentVariables = append(scm.proto.EnvironmentVariables, &proto.EnvironmentVariable{
				Name: envVar,
			})
		}
		for _, secret := range scm.Config.AllSecrets() {
			scm.proto.Secrets = append(scm.proto.Secrets, &proto.Secret{
				Name: secret,
			})
		}
	}

	// Only configure a default API if:
	//  - there is no config or the useDefaultApi config value is true, and
	//  - 'Api' has not already been defined in the schema
	if scm.Config == nil || scm.Config.DefaultApi() {
		defaultApiOverridden := false
		for _, api := range scm.proto.GetApis() {
			if api.GetName() == parser.DefaultApi {
				defaultApiOverridden = true
			}
		}

		if !defaultApiOverridden {
			scm.proto.Apis = append(scm.proto.Apis, defaultAPI(scm.proto))
		}
	}

	// Generate the input messages for all subscribers in the schema.
	scm.makeSubscriberInputMessages()

	return scm.proto
}

// defaultAPI creates an API with all the models and their actions included.
func defaultAPI(scm *proto.Schema) *proto.Api {
	var apiModels []*proto.ApiModel

	for _, m := range scm.GetModels() {
		apiModel := &proto.ApiModel{
			ModelName:    m.GetName(),
			ModelActions: []*proto.ApiModelAction{},
		}

		for _, a := range m.GetActions() {
			apiModel.ModelActions = append(apiModel.ModelActions, &proto.ApiModelAction{ActionName: a.GetName()})
		}

		apiModels = append(apiModels, apiModel)
	}

	return &proto.Api{
		Name:      parser.DefaultApi,
		ApiModels: apiModels,
	}
}

func makeIDQueryInputMessage(name string, modelName *wrapperspb.StringValue) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type:      proto.Type_TYPE_ID,
				ModelName: modelName,
			},
		},
		{
			MessageName: name,
			Name:        "oneOf",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type:      proto.Type_TYPE_ID,
				Repeated:  true,
				ModelName: modelName,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type:      proto.Type_TYPE_ID,
				ModelName: modelName,
			},
		},
	}}
}

func makeStringQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		},
		{
			MessageName: name,
			Name:        "startsWith",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		},
		{
			MessageName: name,
			Name:        "endsWith",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		},
		{
			MessageName: name,
			Name:        "contains",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		},
		{
			MessageName: name,
			Name:        "oneOf",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type:     proto.Type_TYPE_STRING,
				Repeated: true,
			},
		},
	}}
}

func makeStringArrayQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		},
	}}
}

func makeIntQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "lessThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "lessThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "oneOf",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type:     proto.Type_TYPE_INT,
				Repeated: true,
			},
		},
	}}
}

func makeIntArrayQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "lessThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "lessThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		},
	}}
}

func makeDecimalQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "lessThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "lessThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "oneOf",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type:     proto.Type_TYPE_DECIMAL,
				Repeated: true,
			},
		},
	}}
}

func makeDecimalArrayQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "lessThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "lessThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DECIMAL,
			},
		},
	}}
}

func makeBooleanQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_BOOL,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_BOOL,
			},
		},
	}}
}

func makeBooleanArrayQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_BOOL,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_BOOL,
			},
		},
	}}
}

func makeDateQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "before",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "onOrBefore",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "after",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "onOrAfter",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "beforeRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
		{
			MessageName: name,
			Name:        "afterRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
		{
			MessageName: name,
			Name:        "equalsRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
	}}
}

func makeDateArrayQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "before",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "onOrBefore",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "after",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "onOrAfter",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DATE,
			},
		},
		{
			MessageName: name,
			Name:        "beforeRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
		{
			MessageName: name,
			Name:        "afterRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
		{
			MessageName: name,
			Name:        "equalsRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
	}}
}

func makeTimestampQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "before",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_TIMESTAMP,
			},
		},
		{
			MessageName: name,
			Name:        "after",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_TIMESTAMP,
			},
		},
		{
			MessageName: name,
			Name:        "beforeRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
		{
			MessageName: name,
			Name:        "afterRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
		{
			MessageName: name,
			Name:        "equalsRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
	}}
}

func makeTimestampArrayQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "before",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_TIMESTAMP,
			},
		},
		{
			MessageName: name,
			Name:        "after",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_TIMESTAMP,
			},
		},
		{
			MessageName: name,
			Name:        "beforeRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
		{
			MessageName: name,
			Name:        "afterRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
		{
			MessageName: name,
			Name:        "equalsRelative",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_RELATIVE_PERIOD,
			},
		},
	}}
}

func makeDurationQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "lessThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "lessThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
	}}
}

func makeDurationArrayQueryInputMessage(name string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "lessThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "lessThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThan",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
		{
			MessageName: name,
			Name:        "greaterThanOrEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_DURATION,
			},
		},
	}}
}

func makeEnumQueryInputMessage(name string, enumName string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Nullable:    true,
			Optional:    true,
			Type: &proto.TypeInfo{
				Type:     proto.Type_TYPE_ENUM,
				EnumName: wrapperspb.String(enumName),
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Nullable:    true,
			Type: &proto.TypeInfo{
				Type:     proto.Type_TYPE_ENUM,
				EnumName: wrapperspb.String(enumName),
			},
		},
		{
			MessageName: name,
			Name:        "oneOf",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type:     proto.Type_TYPE_ENUM,
				EnumName: wrapperspb.String(enumName),
				Repeated: true,
			},
		},
	}}
}

func makeEnumArrayQueryInputMessage(name string, enumName string) *proto.Message {
	return &proto.Message{Name: name, Fields: []*proto.MessageField{
		{
			MessageName: name,
			Name:        "equals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type:     proto.Type_TYPE_ENUM,
				EnumName: wrapperspb.String(enumName),
			},
		},
		{
			MessageName: name,
			Name:        "notEquals",
			Optional:    true,
			Type: &proto.TypeInfo{
				Type:     proto.Type_TYPE_ENUM,
				EnumName: wrapperspb.String(enumName),
			},
		},
	}}
}

func (scm *Builder) makeListQueryInputMessage(typeInfo *proto.TypeInfo) (*proto.Message, error) {
	var prefix string
	switch typeInfo.GetType() {
	case proto.Type_TYPE_ID:
		prefix = "ID"
	case proto.Type_TYPE_STRING:
		prefix = "String"
	case proto.Type_TYPE_INT:
		prefix = "Int"
	case proto.Type_TYPE_DECIMAL:
		prefix = "Decimal"
	case proto.Type_TYPE_BOOL:
		prefix = "Boolean"
	case proto.Type_TYPE_DATE:
		prefix = "Date"
	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		prefix = "Timestamp"
	case proto.Type_TYPE_DURATION:
		prefix = "Duration"
	case proto.Type_TYPE_ENUM:
		prefix = typeInfo.GetEnumName().GetValue()
	}

	if typeInfo.GetRepeated() {
		msgName := makeInputMessageName(fmt.Sprintf("%sArrayQuery", prefix))

		var enumName *wrapperspb.StringValue
		if typeInfo.GetType() == proto.Type_TYPE_ENUM {
			enumName = typeInfo.GetEnumName()
		}

		var allQueryMsg *proto.Message
		var anyQueryMsg *proto.Message

		allQueryMsgName := makeInputMessageName(fmt.Sprintf("%sArrayAllQuery", prefix))
		anyQueryMsgName := makeInputMessageName(fmt.Sprintf("%sArrayAnyQuery", prefix))

		switch typeInfo.GetType() {
		case proto.Type_TYPE_ID:

		case proto.Type_TYPE_STRING:
			allQueryMsg = makeStringArrayQueryInputMessage(allQueryMsgName)
			anyQueryMsg = makeStringArrayQueryInputMessage(anyQueryMsgName)
		case proto.Type_TYPE_INT:
			allQueryMsg = makeIntArrayQueryInputMessage(allQueryMsgName)
			anyQueryMsg = makeIntArrayQueryInputMessage(anyQueryMsgName)
		case proto.Type_TYPE_DECIMAL:
			allQueryMsg = makeDecimalArrayQueryInputMessage(allQueryMsgName)
			anyQueryMsg = makeDecimalArrayQueryInputMessage(anyQueryMsgName)
		case proto.Type_TYPE_BOOL:
			allQueryMsg = makeBooleanArrayQueryInputMessage(allQueryMsgName)
			anyQueryMsg = makeBooleanArrayQueryInputMessage(anyQueryMsgName)
		case proto.Type_TYPE_DATE:
			allQueryMsg = makeDateArrayQueryInputMessage(allQueryMsgName)
			anyQueryMsg = makeDateArrayQueryInputMessage(anyQueryMsgName)
		case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
			allQueryMsg = makeTimestampArrayQueryInputMessage(allQueryMsgName)
			anyQueryMsg = makeTimestampArrayQueryInputMessage(anyQueryMsgName)
		case proto.Type_TYPE_DURATION:
			allQueryMsg = makeDurationArrayQueryInputMessage(allQueryMsgName)
			anyQueryMsg = makeDurationArrayQueryInputMessage(anyQueryMsgName)
		case proto.Type_TYPE_ENUM:
			allQueryMsg = makeEnumArrayQueryInputMessage(allQueryMsgName, typeInfo.GetEnumName().GetValue())
			anyQueryMsg = makeEnumArrayQueryInputMessage(anyQueryMsgName, typeInfo.GetEnumName().GetValue())
		default:
			return nil, fmt.Errorf("unsupported array query type %s", typeInfo.GetType().String())
		}

		scm.proto.Messages = append(scm.proto.Messages, allQueryMsg)
		scm.proto.Messages = append(scm.proto.Messages, anyQueryMsg)

		return &proto.Message{Name: msgName, Fields: []*proto.MessageField{
			{
				MessageName: msgName,
				Name:        "equals",
				Optional:    true,
				Nullable:    true,
				Type: &proto.TypeInfo{
					Type:     typeInfo.GetType(),
					EnumName: enumName,
					Repeated: true,
				},
			},
			{
				MessageName: msgName,
				Name:        "notEquals",
				Optional:    true,
				Nullable:    true,
				Type: &proto.TypeInfo{
					Type:     typeInfo.GetType(),
					EnumName: enumName,
					Repeated: true,
				},
			},
			{
				MessageName: msgName,
				Name:        "any",
				Optional:    true,
				Type: &proto.TypeInfo{
					Type:        proto.Type_TYPE_MESSAGE,
					MessageName: wrapperspb.String(anyQueryMsgName),
				},
			},
			{
				MessageName: msgName,
				Name:        "all",
				Optional:    true,
				Type: &proto.TypeInfo{
					Type:        proto.Type_TYPE_MESSAGE,
					MessageName: wrapperspb.String(allQueryMsgName),
				},
			},
		}}, nil
	}

	msgName := makeInputMessageName(fmt.Sprintf("%sQuery", prefix))

	switch typeInfo.GetType() {
	case proto.Type_TYPE_ID:
		return makeIDQueryInputMessage(msgName, typeInfo.GetModelName()), nil
	case proto.Type_TYPE_STRING:
		return makeStringQueryInputMessage(msgName), nil
	case proto.Type_TYPE_INT:
		return makeIntQueryInputMessage(msgName), nil
	case proto.Type_TYPE_DECIMAL:
		return makeDecimalQueryInputMessage(msgName), nil
	case proto.Type_TYPE_BOOL:
		return makeBooleanQueryInputMessage(msgName), nil
	case proto.Type_TYPE_DATE:
		return makeDateQueryInputMessage(msgName), nil
	case proto.Type_TYPE_DATETIME, proto.Type_TYPE_TIMESTAMP:
		return makeTimestampQueryInputMessage(msgName), nil
	case proto.Type_TYPE_DURATION:
		return makeDurationQueryInputMessage(msgName), nil
	case proto.Type_TYPE_ENUM:
		return makeEnumQueryInputMessage(msgName, typeInfo.GetEnumName().GetValue()), nil
	default:
		return nil, fmt.Errorf("unsupported query type %s", typeInfo.GetType().String())
	}
}

func makeListOrderByMessages(actionName string, fieldNames []string) []*proto.Message {
	messages := []*proto.Message{}

	for _, fieldName := range fieldNames {
		message := &proto.Message{
			Name:   makeOrderByMessageName(actionName, fieldName),
			Fields: []*proto.MessageField{},
		}

		message.Fields = append(message.Fields, &proto.MessageField{
			MessageName: message.GetName(),
			Name:        fieldName,
			Optional:    false,
			Nullable:    false,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_SORT_DIRECTION,
			},
		})

		messages = append(messages, message)
	}

	return messages
}

// Creates a proto.Message from a slice of action inputs.
func (scm *Builder) makeMessageFromActionInputNodes(name string, inputs []*parser.ActionInputNode, model *parser.ModelNode) *proto.Message {
	fields := []*proto.MessageField{}
	for _, input := range inputs {
		typeInfo, target, targetsOptionalField := scm.inferParserInputType(model, input)

		fields = append(fields, &proto.MessageField{
			Name:        input.Name(),
			Type:        typeInfo,
			Target:      target,
			Optional:    input.Optional,
			Nullable:    targetsOptionalField,
			MessageName: name,
		})
	}

	return &proto.Message{
		Name:   name,
		Fields: fields,
	}
}

// Creates the message structure from an implicit input. For relationships, this will create a nested hierarchy of messages.
func (scm *Builder) makeMessageHierarchyFromImplicitInput(rootMessage *proto.Message, input *parser.ActionInputNode, model *parser.ModelNode, action *parser.ActionNode) {
	target := lo.Map(input.Type.Fragments, func(ident *parser.IdentFragment, _ int) string {
		return ident.Fragment
	})

	currMessage := rootMessage
	currModel := model.Name.Value

	for currIndex, fragment := range target {
		if currIndex < len(target)-1 {
			// If this is not the last target fragment, then we know the current fragment is referring to a related model field.
			// Therefore, we must create a new message for this related model and add it to the current message as a field (if this hasn't already been done with a previous input).

			// Message name of nested message appended with the target framements. E.g. CreateSaleItemsInput
			relatedModelMessageName := makeInputMessageName(action.Name.Value, target[0:currIndex+1]...)

			// Does the field already exist from a previous input?
			fieldAlreadyCreated := false
			for _, f := range currMessage.GetFields() {
				if f.GetName() == fragment {
					fieldAlreadyCreated = true
				}
			}

			// Get field on the model.
			field := query.ModelField(query.Model(scm.asts, currModel), fragment)
			if field == nil {
				panic(fmt.Sprintf("cannot find field %s on model %s", fragment, currModel))
			}

			if !fieldAlreadyCreated {
				// Add the related model message as a field to the current message with typeInfo of Type_TYPE_MESSAGE.
				currMessage.Fields = append(currMessage.Fields, &proto.MessageField{
					Name: fragment,
					Type: &proto.TypeInfo{
						Type: proto.Type_TYPE_MESSAGE,
						// Repeated with be true in a 1:M relationship for create only.
						Repeated: action.Type.Value != parser.ActionTypeList && field.Repeated,
						MessageName: &wrapperspb.StringValue{
							Value: relatedModelMessageName,
						},
					},
					Optional: input.Optional,
					// List op implicit inputs are not nullable, because they will have a query type.
					Nullable:    action.Type.Value != parser.ActionTypeList && field.Optional,
					MessageName: currMessage.GetName(),
				})

				currMessage = &proto.Message{
					Name:   relatedModelMessageName,
					Fields: []*proto.MessageField{},
				}
				scm.proto.Messages = append(scm.proto.Messages, currMessage)
			} else {
				for _, m := range scm.proto.GetMessages() {
					if m.GetName() == relatedModelMessageName {
						currMessage = m
					}
				}
			}

			currModel = field.Type.Value
		} else {
			typeInfo, target, targetsOptionalField := scm.inferParserInputType(model, input)

			if action.Type.Value == parser.ActionTypeList {
				queryMessage, err := scm.makeListQueryInputMessage(typeInfo)
				if err != nil {
					panic(err.Error())
				}

				if !lo.SomeBy(scm.proto.GetMessages(), func(m *proto.Message) bool { return m.GetName() == queryMessage.GetName() }) {
					scm.proto.Messages = append(scm.proto.Messages, queryMessage)
				}

				currMessage.Fields = append(currMessage.Fields, &proto.MessageField{
					Name: fragment,
					Type: &proto.TypeInfo{
						Type:        proto.Type_TYPE_MESSAGE,
						MessageName: wrapperspb.String(queryMessage.GetName())},
					Target:      target,
					Optional:    input.Optional,
					Nullable:    false,
					MessageName: currMessage.GetName(),
				})
			} else {
				// If this is the last or only target, then we add the field to the current message using the typeInfo.
				currMessage.Fields = append(currMessage.Fields, &proto.MessageField{
					Name:        fragment,
					Type:        typeInfo,
					Target:      target,
					Optional:    input.Optional,
					Nullable:    targetsOptionalField,
					MessageName: currMessage.GetName(),
				})
			}
		}
	}
}

// Adds a set of proto.Messages to top level Messages registry for all inputs of an Action.
func (scm *Builder) makeActionInputMessages(model *parser.ModelNode, action *parser.ActionNode) string {
	switch action.Type.Value {
	case parser.ActionTypeCreate:
		if len(action.With) > 0 {
			rootMessage := &proto.Message{
				Name:   makeInputMessageName(action.Name.Value),
				Fields: []*proto.MessageField{},
			}
			scm.proto.Messages = append(scm.proto.Messages, rootMessage)

			for _, input := range action.With {
				if input.Label == nil {
					// If its an implicit input, then create a nested object input structure.
					scm.makeMessageHierarchyFromImplicitInput(rootMessage, input, model, action)
				} else {
					// This is an explicit input, so the first and only fragment will reference the type used.
					typeInfo := scm.explicitInputToTypeInfo(input)

					rootMessage.Fields = append(rootMessage.Fields, &proto.MessageField{
						Name:        input.Label.Value,
						Type:        typeInfo,
						Optional:    input.Optional,
						Nullable:    false, // TODO: can explicit inputs use the null value?
						MessageName: rootMessage.GetName(),
					})
				}
			}

			return rootMessage.GetName()
		}
	case parser.ActionTypeGet, parser.ActionTypeDelete, parser.ActionTypeRead, parser.ActionTypeWrite:
		if len(action.Inputs) > 0 {
			// Create message and add it to the proto schema
			messageName := makeInputMessageName(action.Name.Value)
			message := scm.makeMessageFromActionInputNodes(messageName, action.Inputs, model)
			scm.proto.Messages = append(scm.proto.Messages, message)

			return messageName
		}
	case parser.ActionTypeUpdate:
		fields := []*proto.MessageField{}

		if len(action.Inputs) > 0 {
			// Create where message and add it to the proto schema
			whereMessageName := makeWhereMessageName(action.Name.Value)
			whereMessage := scm.makeMessageFromActionInputNodes(whereMessageName, action.Inputs, model)

			scm.proto.Messages = append(scm.proto.Messages, whereMessage)

			fields = append(fields, &proto.MessageField{
				Name: "where",
				Optional: lo.EveryBy(whereMessage.GetFields(), func(f *proto.MessageField) bool {
					return f.GetOptional()
				}),
				MessageName: makeInputMessageName(action.Name.Value),
				Type: &proto.TypeInfo{
					Type:        proto.Type_TYPE_MESSAGE,
					MessageName: wrapperspb.String(makeWhereMessageName(action.Name.Value)),
				},
			})
		}

		if len(action.With) > 0 {
			// Create values message and add it to the proto schema
			valuesMessage := &proto.Message{
				Name:   makeValuesMessageName(action.Name.Value),
				Fields: []*proto.MessageField{},
			}

			for _, input := range action.With {
				if input.Label == nil {
					// If its an implicit input, then create a nested object input structure.
					scm.makeMessageHierarchyFromImplicitInput(valuesMessage, input, model, action)
				} else {
					// This is an explicit input, so the first and only fragment will reference the type used.
					typeInfo := scm.explicitInputToTypeInfo(input)

					valuesMessage.Fields = append(valuesMessage.Fields, &proto.MessageField{
						Name:        input.Label.Value,
						Type:        typeInfo,
						Optional:    input.Optional,
						Nullable:    false, // TODO: can explicit inputs use the null value?
						MessageName: valuesMessage.GetName(),
					})
				}
			}

			scm.proto.Messages = append(scm.proto.Messages, valuesMessage)

			fields = append(fields, &proto.MessageField{
				Name: "values",
				Optional: lo.EveryBy(valuesMessage.GetFields(), func(f *proto.MessageField) bool {
					return f.GetOptional()
				}),
				MessageName: makeInputMessageName(action.Name.Value),
				Type: &proto.TypeInfo{
					Type:        proto.Type_TYPE_MESSAGE,
					MessageName: wrapperspb.String(makeValuesMessageName(action.Name.Value)),
				},
			})
		}

		if len(fields) > 0 {
			// Create root action message with "where" and "values" fields.
			scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
				Name:   makeInputMessageName(action.Name.Value),
				Fields: fields,
			})

			return makeInputMessageName(action.Name.Value)
		}
	case parser.ActionTypeList:

		whereMessage := &proto.Message{
			Name:   makeWhereMessageName(action.Name.Value),
			Fields: []*proto.MessageField{},
		}

		for _, input := range action.Inputs {
			if input.Label == nil {
				scm.makeMessageHierarchyFromImplicitInput(whereMessage, input, model, action)
			} else {
				typeInfo := scm.explicitInputToTypeInfo(input)

				whereMessage.Fields = append(whereMessage.Fields, &proto.MessageField{
					Name:        input.Name(),
					Type:        typeInfo,
					Optional:    input.Optional,
					MessageName: makeWhereMessageName(action.Name.Value),
				})
			}
		}

		sortableFields, err := query.ActionSortableFieldNames(action)
		if err != nil {
			panic(err)
		}

		inputMessage := &proto.Message{
			Name:   makeInputMessageName(action.Name.Value),
			Fields: []*proto.MessageField{},
		}

		// Only add where field if there are inputs
		if len(action.Inputs) > 0 {
			scm.proto.Messages = append(scm.proto.Messages, whereMessage)

			inputMessage.Fields = append(inputMessage.Fields, &proto.MessageField{
				Name: "where",
				Optional: lo.EveryBy(whereMessage.GetFields(), func(f *proto.MessageField) bool {
					return f.GetOptional()
				}),
				MessageName: makeInputMessageName(action.Name.Value),
				Type: &proto.TypeInfo{
					Type:        proto.Type_TYPE_MESSAGE,
					MessageName: wrapperspb.String(whereMessage.GetName()),
				},
			})
		}

		// Include pagination fields
		inputMessage.Fields = append(inputMessage.Fields, &proto.MessageField{
			Name:        "first",
			MessageName: makeInputMessageName(action.Name.Value),
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		}, &proto.MessageField{
			Name:        "after",
			MessageName: makeInputMessageName(action.Name.Value),
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		}, &proto.MessageField{
			Name:        "last",
			MessageName: makeInputMessageName(action.Name.Value),
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		}, &proto.MessageField{
			Name:        "before",
			MessageName: makeInputMessageName(action.Name.Value),
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_STRING,
			},
		}, &proto.MessageField{
			Name:        "limit",
			MessageName: makeInputMessageName(action.Name.Value),
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		}, &proto.MessageField{
			Name:        "offset",
			MessageName: makeInputMessageName(action.Name.Value),
			Optional:    true,
			Type: &proto.TypeInfo{
				Type: proto.Type_TYPE_INT,
			},
		})

		orderByMessages := makeListOrderByMessages(action.Name.Value, sortableFields)
		if len(orderByMessages) > 0 {
			orderByMessageField := &proto.MessageField{
				Name:        "orderBy",
				MessageName: makeInputMessageName(action.Name.Value),
				Optional:    true,
				Type: &proto.TypeInfo{
					Type:       proto.Type_TYPE_UNION,
					Repeated:   true,
					UnionNames: lo.Map(orderByMessages, func(m *proto.Message, _ int) *wrapperspb.StringValue { return wrapperspb.String(m.GetName()) }),
				},
			}

			scm.proto.Messages = append(scm.proto.Messages, orderByMessages...)
			inputMessage.Fields = append(inputMessage.Fields, orderByMessageField)
		}

		scm.proto.Messages = append(scm.proto.Messages, inputMessage)

		return inputMessage.GetName()
	default:
		panic("unhandled action type when creating input message types")
	}

	return ""
}

func (scm *Builder) makeModel(decl *parser.DeclarationNode) {
	parserModel := decl.Model
	protoModel := &proto.Model{
		Name: parserModel.Name.Value,
	}

	for _, section := range parserModel.Sections {
		switch {
		case section.Fields != nil:
			fields := scm.makeFields(section.Fields, protoModel.GetName())
			protoModel.Fields = append(protoModel.Fields, fields...)
		case section.Actions != nil:
			ops := scm.makeActions(section.Actions, protoModel.GetName(), parserModel.BuiltIn)
			protoModel.Actions = append(protoModel.Actions, ops...)
		case section.Attribute != nil:
			scm.applyModelAttribute(parserModel, protoModel, section.Attribute)
		default:
			// this is possible if the user defines an empty block in the schema e.g. "fields {}"
			// this isn't really an error so we can just ignore these sections
		}
	}

	scm.proto.Models = append(scm.proto.Models, protoModel)
}

func (scm *Builder) makeRole(decl *parser.DeclarationNode) {
	parserRole := decl.Role
	protoRole := &proto.Role{
		Name: parserRole.Name.Value,
	}
	for _, section := range parserRole.Sections {
		for _, parserDomain := range section.Domains {
			protoRole.Domains = append(protoRole.Domains, stripQuotes(parserDomain.Domain))
		}
		for _, parserEmail := range section.Emails {
			protoRole.Emails = append(protoRole.Emails, stripQuotes(parserEmail.Email))
		}
	}
	scm.proto.Roles = append(scm.proto.Roles, protoRole)
}

func (scm *Builder) makeAPI(decl *parser.DeclarationNode) {
	parserAPI := decl.API
	protoAPI := &proto.Api{
		Name:      parserAPI.Name.Value,
		ApiModels: []*proto.ApiModel{},
	}
	for _, section := range parserAPI.Sections {
		for _, parserApiModel := range section.Models {
			protoModel := &proto.ApiModel{
				ModelName:    parserApiModel.Name.Value,
				ModelActions: []*proto.ApiModelAction{},
			}

			if len(parserApiModel.Sections) == 0 {
				model := query.Model(scm.asts, parserApiModel.Name.Value)
				actions := query.ModelActions(model)

				if model != nil {
					for _, a := range actions {
						protoModel.ModelActions = append(protoModel.ModelActions, &proto.ApiModelAction{ActionName: a.Name.Value})
					}
				}
			} else {
				for _, a := range parserApiModel.Sections[0].Actions {
					protoModel.ModelActions = append(protoModel.ModelActions, &proto.ApiModelAction{ActionName: a.Name.Value})
				}
			}

			protoAPI.ApiModels = append(protoAPI.ApiModels, protoModel)
		}
	}
	scm.proto.Apis = append(scm.proto.Apis, protoAPI)
}

func (scm *Builder) makeAnyType() {
	anyMsg := &proto.Message{
		Name: "Any",
	}

	scm.proto.Messages = append(scm.proto.Messages, anyMsg)
}

func (scm *Builder) makeMessage(decl *parser.DeclarationNode) {
	parserMsg := decl.Message

	fields := lo.Map(parserMsg.Fields, func(f *parser.FieldNode, _ int) *proto.MessageField {
		field := &proto.MessageField{
			Name: f.Name.Value,
			Type: &proto.TypeInfo{
				Type:     scm.parserTypeToProtoType(f.Type.Value),
				Repeated: f.Repeated,
			},
			Optional:    f.Optional,
			MessageName: parserMsg.Name.Value,
		}

		if field.GetType().GetType() == proto.Type_TYPE_ENUM {
			field.Type.EnumName = wrapperspb.String(f.Type.Value)
		}

		if field.GetType().GetType() == proto.Type_TYPE_MESSAGE {
			field.Type.MessageName = wrapperspb.String(f.Type.Value)
		}

		if field.GetType().GetType() == proto.Type_TYPE_MODEL {
			field.Type.ModelName = wrapperspb.String(f.Type.Value)
		}

		return field
	})

	scm.proto.Messages = append(scm.proto.Messages, &proto.Message{
		Name:   parserMsg.Name.Value,
		Fields: fields,
	})
}

func (scm *Builder) makeEnum(decl *parser.DeclarationNode) {
	parserEnum := decl.Enum
	enum := &proto.Enum{
		Name:   parserEnum.Name.Value,
		Values: []*proto.EnumValue{},
	}
	for _, value := range parserEnum.Values {
		enum.Values = append(enum.Values, &proto.EnumValue{
			Name: value.Name.Value,
		})
	}
	scm.proto.Enums = append(scm.proto.Enums, enum)
}

func (scm *Builder) makeJob(decl *parser.DeclarationNode) {
	parserJob := decl.Job
	messageName := makeMessageName(parserJob.Name.Value)

	job := &proto.Job{
		Name: parserJob.Name.Value,
	}

	message := &proto.Message{
		Name:   messageName,
		Fields: []*proto.MessageField{},
	}

	for _, section := range parserJob.Sections {
		switch {
		case section.Attribute != nil:
			scm.applyJobAttribute(job, section.Attribute)
		case section.Inputs != nil:
			scm.applyJobInputs(message, section.Inputs)
		default:
			panic(fmt.Sprintf("unhandled section when parsing job '%s'", job.GetName()))
		}
	}

	if len(message.GetFields()) > 0 {
		job.InputMessageName = message.GetName()
		scm.proto.Messages = append(scm.proto.Messages, message)
	}

	scm.proto.Jobs = append(scm.proto.Jobs, job)
}

func (scm *Builder) makeRoutes(decl *parser.DeclarationNode) {
	for _, route := range decl.Routes.Routes {
		var method proto.HttpMethod
		switch strings.ToUpper(route.Method.Value) {
		case http.MethodGet:
			method = proto.HttpMethod_HTTP_METHOD_GET
		case http.MethodPost:
			method = proto.HttpMethod_HTTP_METHOD_POST
		case http.MethodPut:
			method = proto.HttpMethod_HTTP_METHOD_PUT
		case http.MethodDelete:
			method = proto.HttpMethod_HTTP_METHOD_DELETE
		}

		pattern := route.Pattern.Value
		pattern = strings.TrimPrefix(pattern, `"`)
		pattern = strings.TrimSuffix(pattern, `"`)
		if !strings.HasPrefix(pattern, "/") {
			pattern = "/" + pattern
		}

		scm.proto.Routes = append(scm.proto.Routes, &proto.Route{
			Method:  method,
			Pattern: pattern,
			Handler: route.Handler.Value,
		})
	}
}

func (scm *Builder) makeFlow(decl *parser.DeclarationNode) {
	parserFlow := decl.Flow
	messageName := makeMessageName(parserFlow.Name.Value)

	flow := &proto.Flow{
		Name: parserFlow.Name.Value,
	}

	message := &proto.Message{
		Name:   messageName,
		Fields: []*proto.MessageField{},
	}

	for _, section := range parserFlow.Sections {
		switch {
		case section.Attribute != nil:
			scm.applyFlowAttribute(flow, section.Attribute)
		case section.Inputs != nil:
			scm.applyFlowInputs(message, section.Inputs)
		default:
			panic(fmt.Sprintf("unhandled section when parsing flow '%s'", flow.GetName()))
		}
	}

	if len(message.GetFields()) > 0 {
		flow.InputMessageName = message.GetName()
		scm.proto.Messages = append(scm.proto.Messages, message)
	}

	scm.proto.Flows = append(scm.proto.Flows, flow)
}

func (scm *Builder) makeFields(parserFields []*parser.FieldNode, modelName string) []*proto.Field {
	protoFields := []*proto.Field{}
	for _, parserField := range parserFields {
		protoField := scm.makeField(parserField, modelName)
		protoFields = append(protoFields, protoField)
	}
	return protoFields
}

func (scm *Builder) makeField(parserField *parser.FieldNode, modelName string) *proto.Field {
	typeInfo := scm.parserFieldToProtoTypeInfo(parserField)
	protoField := &proto.Field{
		ModelName: modelName,
		Name:      parserField.Name.Value,
		Type:      typeInfo,
		Optional:  parserField.Optional,
	}

	// Handle @unique attribute at model level which expresses
	// unique constrains across multiple fields
	model := query.Model(scm.asts, modelName)
	for _, attr := range query.ModelAttributes(model) {
		if attr.Name.Value != parser.AttributeUnique {
			continue
		}

		idents, _ := resolve.AsIdentArray(attr.Arguments[0].Expression)
		fieldNames := lo.Map(idents, func(v *parser.ExpressionIdent, i int) string {
			return v.String()
		})

		if !lo.Contains(fieldNames, parserField.Name.Value) {
			continue
		}

		protoField.UniqueWith = lo.Filter(fieldNames, func(v string, i int) bool {
			return v != parserField.Name.Value
		})
	}

	scm.applyFieldAttributes(parserField, protoField)

	// Auto-inserted foreign key field
	if query.IsForeignKey(scm.asts, model, parserField) {
		modelField := query.Field(model, strings.TrimSuffix(parserField.Name.Value, "Id"))
		protoField.ForeignKeyInfo = &proto.ForeignKeyInfo{
			RelatedModelName:  modelField.Type.Value,
			RelatedModelField: parser.FieldNameId,
		}
	}

	relationship, err := query.GetRelationship(scm.asts, query.Model(scm.asts, modelName), parserField)
	if err != nil {
		panic(err)
	}
	if relationship != nil {
		if relationship.Field == nil ||
			query.ValidOneToHasMany(parserField, relationship.Field) ||
			query.ValidUniqueOneToHasOne(parserField, relationship.Field) {
			protoField.ForeignKeyFieldName = wrapperspb.String(fmt.Sprintf("%sId", parserField.Name.Value))
		}
	}

	// If this is a HasMany or BelongsTo relationship field - see if we can mark it with
	// an explicit InverseFieldName - i.e. one defined by an @relation attribute.
	if protoField.GetType().GetType() == proto.Type_TYPE_MODEL {
		scm.setInverseFieldName(parserField, protoField)
	}

	return protoField
}

// setInverseFieldName works on fields of type Model that are repeated. It looks to
// see if the schema defines an explicit inverse relationship field for it, and when so, sets
// this field's InverseFieldName property accordingly.
func (scm *Builder) setInverseFieldName(thisParserField *parser.FieldNode, thisProtoField *proto.Field) {
	// We have to look in the related model's fields, to see if any of them have an @relation
	// attribute that refers back to this field.

	nameOfRelatedModel := thisProtoField.GetType().GetModelName().GetValue()
	relatedModel := query.Model(scm.asts, nameOfRelatedModel)

	// Use the field name in @relation(fieldName) if this attribute exists
	relationAttr := query.FieldGetAttribute(thisParserField, parser.AttributeRelation)
	if relationAttr != nil {
		inverseFieldName := attributeFirstArgAsIdentifier(relationAttr)
		thisProtoField.InverseFieldName = wrapperspb.String(inverseFieldName)
		return
	}

	// If no @relation attribute exists, then look for a match in the related model fields' @relation attributes
	for _, remoteField := range query.ModelFields(relatedModel) {
		if remoteField.Type.Value != thisProtoField.GetModelName() {
			continue
		}
		relationAttr := query.FieldGetAttribute(remoteField, parser.AttributeRelation)
		if relationAttr != nil {
			inverseFieldName := attributeFirstArgAsIdentifier(relationAttr)
			if inverseFieldName == thisProtoField.GetName() {
				thisProtoField.InverseFieldName = wrapperspb.String(remoteField.Name.Value)
				return
			}
		}
	}

	// If there are no @relation attributes that match, then we know that there is only one relation
	// between these models of this exact relationship type and in this direction
	for _, remoteField := range query.ModelFields(relatedModel) {
		if remoteField.Type.Value != thisProtoField.GetModelName() {
			continue
		}
		if nameOfRelatedModel == thisProtoField.GetModelName() && remoteField.Name.Value == thisProtoField.GetName() {
			continue
		}
		if query.ValidOneToHasMany(thisParserField, remoteField) ||
			query.ValidOneToHasMany(remoteField, thisParserField) ||
			query.ValidUniqueOneToHasOne(thisParserField, remoteField) ||
			query.ValidUniqueOneToHasOne(remoteField, thisParserField) {
			thisProtoField.InverseFieldName = wrapperspb.String(remoteField.Name.Value)
		}
	}
}

// attributeFirstArgAsIdentifier fishes out the identifier being held
// by the first argument of the given attribute. It must only be called when
// you know that it is well formed for that.
func attributeFirstArgAsIdentifier(attr *parser.AttributeNode) string {
	expr := attr.Arguments[0].Expression
	theString, _ := resolve.AsIdent(expr)
	return theString.String()
}

func (scm *Builder) makeActions(actions []*parser.ActionNode, modelName string, builtIn bool) []*proto.Action {
	protoOps := []*proto.Action{}

	for _, action := range actions {
		protoOp := scm.makeAction(action, modelName, builtIn)
		protoOps = append(protoOps, protoOp)
	}
	return protoOps
}

func (scm *Builder) makeAction(action *parser.ActionNode, modelName string, builtIn bool) *proto.Action {
	var implementation proto.ActionImplementation

	switch {
	case builtIn:
		implementation = proto.ActionImplementation_ACTION_IMPLEMENTATION_RUNTIME
	case action.IsFunction():
		implementation = proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM
	default:
		implementation = proto.ActionImplementation_ACTION_IMPLEMENTATION_AUTO
	}

	protoAction := &proto.Action{
		ModelName:      modelName,
		Name:           action.Name.Value,
		Implementation: implementation,
		Type:           scm.mapToActionType(action.Type.Value),
	}

	model := query.Model(scm.asts, modelName)

	if action.IsArbitraryFunction() {
		// Does the function have any inputs defined?
		if action.Inputs != nil {
			// if its an arbitrary function, then the input will exist in scm.Messages unless the inputs were defined inline
			// output messages will always be defined in scm.Messages
			usesAny := action.Inputs[0].Type.ToString() == parser.MessageFieldTypeAny
			usingInlineInputs := true

			for _, ast := range scm.asts {
				for _, d := range ast.Declarations {
					if d.Message != nil && d.Message.Name.Value == action.Inputs[0].Type.ToString() {
						usingInlineInputs = false
					}
				}
			}

			switch {
			case usesAny:
				protoAction.InputMessageName = action.Inputs[0].Type.ToString()
			case usingInlineInputs:
				protoAction.InputMessageName = scm.makeActionInputMessages(model, action)
			default:
				protoAction.InputMessageName = action.Inputs[0].Type.ToString()
			}
		}

		protoAction.ResponseMessageName = action.Returns[0].Type.ToString()
	} else {
		// we need to generate the messages representing the inputs to the scm.Messages
		protoAction.InputMessageName = scm.makeActionInputMessages(model, action)
	}

	scm.applyActionAttributes(action, protoAction, modelName)

	return protoAction
}

func (scm *Builder) inferParserInputType(
	model *parser.ModelNode,
	input *parser.ActionInputNode,
) (t *proto.TypeInfo, target []string, targetsOptionalField bool) {
	idents := input.Type.Fragments
	protoType := scm.parserTypeToProtoType(idents[0].Fragment)

	var modelName *wrapperspb.StringValue
	var fieldName *wrapperspb.StringValue
	var enumName *wrapperspb.StringValue
	repeated := false

	if protoType == proto.Type_TYPE_ENUM {
		enumName = &wrapperspb.StringValue{
			Value: idents[0].Fragment,
		}
	}

	targetsOptionalField = false

	if protoType == proto.Type_TYPE_UNKNOWN {
		// If we haven't been able to resolve the type of the input it
		// must be a model field, so we need to resolve it

		var field *parser.FieldNode
		currModel := model

		for _, ident := range idents {
			target = append(target, ident.Fragment)

			field = query.ModelField(currModel, ident.Fragment)

			m := query.Model(scm.asts, field.Type.Value)
			if m != nil {
				currModel = m
			}
		}

		if field != nil && field.Optional {
			targetsOptionalField = true
		}

		protoType = scm.parserFieldToProtoTypeInfo(field).GetType()

		modelName = &wrapperspb.StringValue{
			Value: currModel.Name.Value,
		}
		fieldName = &wrapperspb.StringValue{
			Value: field.Name.Value,
		}

		repeated = field.Repeated

		if protoType == proto.Type_TYPE_ENUM {
			enumName = &wrapperspb.StringValue{
				Value: field.Type.Value,
			}
		}
	}

	return &proto.TypeInfo{
		Type:      protoType,
		Repeated:  repeated,
		ModelName: modelName,
		FieldName: fieldName,
		EnumName:  enumName,
	}, target, targetsOptionalField
}

// parserType could be a built-in type or a user-defined model or enum.
func (scm *Builder) parserTypeToProtoType(parserType string) proto.Type {
	switch {
	case parserType == parser.FieldTypeText:
		return proto.Type_TYPE_STRING
	case parserType == parser.FieldTypeID:
		return proto.Type_TYPE_ID
	case parserType == parser.FieldTypeBoolean:
		return proto.Type_TYPE_BOOL
	case parserType == parser.FieldTypeNumber:
		return proto.Type_TYPE_INT
	case parserType == parser.FieldTypeDate:
		return proto.Type_TYPE_DATE
	case parserType == parser.FieldTypeTimestamp:
		return proto.Type_TYPE_DATETIME
	case parserType == parser.FieldTypeSecret:
		return proto.Type_TYPE_SECRET
	case parserType == parser.FieldTypePassword:
		return proto.Type_TYPE_PASSWORD
	case query.IsModel(scm.asts, parserType):
		return proto.Type_TYPE_MODEL
	case query.IsEnum(scm.asts, parserType):
		return proto.Type_TYPE_ENUM
	case query.IsMessage(scm.asts, parserType):
		return proto.Type_TYPE_MESSAGE
	case parserType == parser.MessageFieldTypeAny:
		return proto.Type_TYPE_ANY
	case parserType == parser.FieldTypeMarkdown:
		return proto.Type_TYPE_MARKDOWN
	case parserType == parser.FieldTypeDecimal:
		return proto.Type_TYPE_DECIMAL
	case parserType == parser.FieldTypeVector:
		return proto.Type_TYPE_VECTOR
	case parserType == parser.FieldTypeFile:
		return proto.Type_TYPE_FILE
	case parserType == parser.FieldTypeDuration:
		return proto.Type_TYPE_DURATION
	default:
		return proto.Type_TYPE_UNKNOWN
	}
}

func (scm *Builder) explicitInputToTypeInfo(input *parser.ActionInputNode) *proto.TypeInfo {
	protoType := scm.parserTypeToProtoType(input.Type.Fragments[0].Fragment)

	disallowedExplicitInputTypes := []proto.Type{
		proto.Type_TYPE_MODEL,
		proto.Type_TYPE_MESSAGE,
		proto.Type_TYPE_ANY,
		proto.Type_TYPE_UNKNOWN}

	if lo.Contains(disallowedExplicitInputTypes, protoType) {
		panic(fmt.Sprintf("explicit input field %s cannot be of type %s", input.Name(), protoType))
	}

	var enumName *wrapperspb.StringValue
	if protoType == proto.Type_TYPE_ENUM {
		enumName = &wrapperspb.StringValue{
			Value: query.Enum(scm.asts, input.Type.Fragments[0].Fragment).Name.Value,
		}
	}

	return &proto.TypeInfo{
		Type:     protoType,
		EnumName: enumName,
	}
}

func (scm *Builder) parserFieldToProtoTypeInfo(field *parser.FieldNode) *proto.TypeInfo {
	protoType := scm.parserTypeToProtoType(field.Type.Value)
	var modelName *wrapperspb.StringValue
	var enumName *wrapperspb.StringValue

	switch protoType {
	case proto.Type_TYPE_MODEL:
		modelName = &wrapperspb.StringValue{
			Value: query.Model(scm.asts, field.Type.Value).Name.Value,
		}
	case proto.Type_TYPE_ENUM:
		enumName = &wrapperspb.StringValue{
			Value: query.Enum(scm.asts, field.Type.Value).Name.Value,
		}
	}

	return &proto.TypeInfo{
		Type:      protoType,
		Repeated:  field.Repeated,
		ModelName: modelName,
		EnumName:  enumName,
	}
}

func (scm *Builder) applyModelAttribute(parserModel *parser.ModelNode, protoModel *proto.Model, attribute *parser.AttributeNode) {
	switch attribute.Name.Value {
	case parser.AttributePermission:
		perm := scm.permissionAttributeToProtoPermission(attribute)
		perm.ModelName = protoModel.GetName()
		protoModel.Permissions = append(protoModel.Permissions, perm)
	case parser.AttributeOn:
		subscriberName, _ := resolve.AsIdent(attribute.Arguments[1].Expression)

		// Create the subscriber if it has not yet been created yet.
		subscriber := proto.FindSubscriber(scm.proto.GetSubscribers(), subscriberName.Fragments[0])
		if subscriber == nil {
			subscriber = &proto.Subscriber{
				Name:             subscriberName.Fragments[0],
				InputMessageName: makeSubscriberMessageName(subscriberName.Fragments[0]),
				EventNames:       []string{},
			}
			scm.proto.Subscribers = append(scm.proto.Subscribers, subscriber)
		}

		// For each event, add to the proto schema if it doesn't exist,
		// and add it to the current subscriber's EventNames field.
		actionTypesArg, _ := resolve.AsIdentArray(attribute.Arguments[0].Expression)
		for _, arg := range actionTypesArg {
			actionType := scm.mapToActionType(arg.Fragments[0])
			eventName := makeEventName(parserModel.Name.Value, mapToEventType(actionType))

			event := proto.FindEvent(scm.proto.GetEvents(), eventName)
			if event == nil {
				event = &proto.Event{
					Name:       eventName,
					ModelName:  parserModel.Name.Value,
					ActionType: actionType,
				}
				scm.proto.Events = append(scm.proto.Events, event)
			}

			subscriber.EventNames = append(subscriber.EventNames, eventName)
		}
	}
}

// makeSubscriberInputMessages creates the event input messages for the subscriber functions.
// The signature of these messages depends on which events the subscriber is handling.
func (scm *Builder) makeSubscriberInputMessages() {
	for _, subscriber := range scm.proto.GetSubscribers() {
		message := &proto.Message{
			Name:   subscriber.GetInputMessageName(),
			Fields: []*proto.MessageField{},
			Type: &proto.TypeInfo{
				Type:       proto.Type_TYPE_UNION,
				UnionNames: []*wrapperspb.StringValue{},
			},
		}

		scm.proto.Messages = append(scm.proto.Messages, message)

		for _, eventName := range subscriber.GetEventNames() {
			event := proto.FindEvent(scm.proto.GetEvents(), eventName)

			eventMessage := &proto.Message{
				Name:   makeSubscriberMessageEventName(subscriber.GetName(), event.GetModelName(), mapToEventType(event.GetActionType())),
				Fields: []*proto.MessageField{},
			}

			eventTargetMessage := &proto.Message{
				Name:   makeSubscriberMessageEventTargetName(subscriber.GetName(), event.GetModelName(), mapToEventType(event.GetActionType())),
				Fields: []*proto.MessageField{},
			}

			eventName := makeEventName(event.GetModelName(), mapToEventType(event.GetActionType()))

			eventMessage.Fields = append(eventMessage.Fields, &proto.MessageField{
				MessageName: eventMessage.GetName(),
				Name:        "eventName",
				Type: &proto.TypeInfo{
					Type:               proto.Type_TYPE_STRING_LITERAL,
					StringLiteralValue: wrapperspb.String(eventName),
				},
			})

			eventMessage.Fields = append(eventMessage.Fields, &proto.MessageField{
				MessageName: eventMessage.GetName(),
				Name:        "occurredAt",
				Type:        &proto.TypeInfo{Type: proto.Type_TYPE_TIMESTAMP},
			})

			eventMessage.Fields = append(eventMessage.Fields, &proto.MessageField{
				MessageName: eventMessage.GetName(),
				Name:        "identityId",
				Optional:    true,
				Type:        &proto.TypeInfo{Type: proto.Type_TYPE_ID},
			})

			eventMessage.Fields = append(eventMessage.Fields, &proto.MessageField{
				MessageName: eventMessage.GetName(),
				Name:        "target",
				Type: &proto.TypeInfo{
					Type:        proto.Type_TYPE_MESSAGE,
					MessageName: wrapperspb.String(eventTargetMessage.GetName()),
				},
			})

			eventTargetMessage.Fields = append(eventTargetMessage.Fields, &proto.MessageField{
				MessageName: eventTargetMessage.GetName(),
				Name:        "id",
				Type:        &proto.TypeInfo{Type: proto.Type_TYPE_ID},
			})

			eventTargetMessage.Fields = append(eventTargetMessage.Fields, &proto.MessageField{
				MessageName: eventTargetMessage.GetName(),
				Name:        "type",
				Type:        &proto.TypeInfo{Type: proto.Type_TYPE_STRING},
			})

			eventTargetMessage.Fields = append(eventTargetMessage.Fields, &proto.MessageField{
				MessageName: eventTargetMessage.GetName(),
				Name:        "data",
				Type: &proto.TypeInfo{
					Type:      proto.Type_TYPE_MODEL,
					ModelName: wrapperspb.String(event.GetModelName()),
				},
			})

			if event.GetActionType() != proto.ActionType_ACTION_TYPE_CREATE {
				eventTargetMessage.Fields = append(eventTargetMessage.Fields, &proto.MessageField{
					MessageName: eventTargetMessage.GetName(),
					Name:        "previousData",
					Type: &proto.TypeInfo{
						Type:      proto.Type_TYPE_MODEL,
						ModelName: wrapperspb.String(event.GetModelName()),
					},
				})
			}

			message.Type.UnionNames = append(message.Type.UnionNames, wrapperspb.String(eventMessage.GetName()))
			scm.proto.Messages = append(scm.proto.Messages, eventMessage)
			scm.proto.Messages = append(scm.proto.Messages, eventTargetMessage)
		}
	}
}

func (scm *Builder) applyActionAttributes(action *parser.ActionNode, protoAction *proto.Action, modelName string) {
	for _, attribute := range action.Attributes {
		switch attribute.Name.Value {
		case parser.AttributePermission:
			perm := scm.permissionAttributeToProtoPermission(attribute)
			perm.ModelName = modelName
			perm.ActionName = wrapperspb.String(protoAction.GetName())
			protoAction.Permissions = append(protoAction.Permissions, perm)
		case parser.AttributeWhere:
			expr := attribute.Arguments[0].Expression.String()
			where := &proto.Expression{Source: expr}
			protoAction.WhereExpressions = append(protoAction.WhereExpressions, where)
		case parser.AttributeSet:
			expr := attribute.Arguments[0].Expression.String()
			set := &proto.Expression{Source: expr}
			protoAction.SetExpressions = append(protoAction.SetExpressions, set)
		case parser.AttributeValidate:
			expr := attribute.Arguments[0].Expression.String()
			set := &proto.Expression{Source: expr}
			protoAction.ValidationExpressions = append(protoAction.ValidationExpressions, set)
		case parser.AttributeEmbed:
			for _, arg := range attribute.Arguments {
				expr := arg.Expression.String()
				protoAction.ResponseEmbeds = append(protoAction.ResponseEmbeds, expr)
			}
		case parser.AttributeOrderBy:
			for _, arg := range attribute.Arguments {
				field := arg.Label.Value
				direction := arg.Expression.String()
				orderBy := &proto.OrderByStatement{
					FieldName: field,
					Direction: mapToOrderByDirection(direction),
				}
				protoAction.OrderBy = append(protoAction.OrderBy, orderBy)
			}
		case parser.AttributeFacet:
			for _, arg := range attribute.Arguments {
				expr := arg.Expression.String()
				protoAction.Facets = append(protoAction.Facets, expr)
			}
		}
	}
}

func (scm *Builder) applyFieldAttributes(parserField *parser.FieldNode, protoField *proto.Field) {
	for _, fieldAttribute := range parserField.Attributes {
		switch fieldAttribute.Name.Value {
		case parser.AttributeUnique:
			protoField.Unique = true
		case parser.AttributeSequence:
			protoField.Unique = true
			prefix, _, _ := resolve.ToValue[string](fieldAttribute.Arguments[0].Expression)
			protoField.Sequence = &proto.Sequence{
				Prefix:   prefix,
				StartsAt: 1,
			}
			if len(fieldAttribute.Arguments) >= 2 {
				startsAt, _, _ := resolve.ToValue[int64](fieldAttribute.Arguments[1].Expression)
				protoField.Sequence.StartsAt = uint32(startsAt)
			}
		case parser.AttributePrimaryKey:
			protoField.PrimaryKey = true
			protoField.Unique = true
		case parser.AttributeDefault:
			defaultValue := &proto.DefaultValue{}
			if len(fieldAttribute.Arguments) == 1 {
				defaultValue.Expression = &proto.Expression{
					Source: fieldAttribute.Arguments[0].Expression.String(),
				}
			} else {
				defaultValue.UseZeroValue = true
			}
			protoField.DefaultValue = defaultValue
		case parser.AttributeComputed:
			protoField.ComputedExpression = &proto.Expression{
				Source: fieldAttribute.Arguments[0].Expression.String(),
			}
		case parser.AttributeRelation:
			// We cannot process this field attribute here. But here is an explanation
			// of why that is so - for future readers.
			//
			// This attribute (the @relation attribute) is placed one HasOne relation fields in the input schema -
			// to specify a field in its related model that is its inverse. We decided this because
			// it seems most intuitive for the user - given that to use 1:many relations at all,
			// you HAVE TO HAVE the hasOne end.
			//
			// HOWEVER we want the InverseFieldName field property in the protobuf representation
			// to live on the RELATED model's field, i.e. the repeated relationship field - NOT this field.
			//
			// The problem is that the related model might not even be present yet in the proto.Schema that is
			// currently under construction - because the call-graph of the construction process builds the proto
			// for each model in turn, and might not have reached the related model yet.
			//
			// INSTEAD we sort it all out when we reach hasMany fields at the other end of the inverse relation.
			// See the call to setExplicitInverseFieldName() at the end of scm.makeField().
		}
	}
}

func (scm *Builder) permissionAttributeToProtoPermission(attr *parser.AttributeNode) *proto.PermissionRule {
	pr := &proto.PermissionRule{}
	for _, arg := range attr.Arguments {
		switch arg.Label.Value {
		case "expression":
			expr := arg.Expression.String()
			pr.Expression = &proto.Expression{Source: expr}
		case "roles":
			idents, _ := resolve.AsIdentArray(arg.Expression)
			for _, item := range idents {
				pr.RoleNames = append(pr.RoleNames, item.Fragments[0])
			}
		case "actions":
			idents, _ := resolve.AsIdentArray(arg.Expression)
			for _, items := range idents {
				pr.ActionTypes = append(pr.ActionTypes, scm.mapToActionType(items.Fragments[0]))
			}
		}
	}
	return pr
}

func (scm *Builder) mapToActionType(actionType string) proto.ActionType {
	switch actionType {
	case parser.ActionTypeCreate:
		return proto.ActionType_ACTION_TYPE_CREATE
	case parser.ActionTypeUpdate:
		return proto.ActionType_ACTION_TYPE_UPDATE
	case parser.ActionTypeGet:
		return proto.ActionType_ACTION_TYPE_GET
	case parser.ActionTypeList:
		return proto.ActionType_ACTION_TYPE_LIST
	case parser.ActionTypeDelete:
		return proto.ActionType_ACTION_TYPE_DELETE
	case parser.ActionTypeRead:
		return proto.ActionType_ACTION_TYPE_READ
	case parser.ActionTypeWrite:
		return proto.ActionType_ACTION_TYPE_WRITE
	default:
		return proto.ActionType_ACTION_TYPE_UNKNOWN
	}
}

func mapToEventType(actionType proto.ActionType) string {
	switch actionType {
	case proto.ActionType_ACTION_TYPE_CREATE:
		return "created"
	case proto.ActionType_ACTION_TYPE_UPDATE:
		return "updated"
	case proto.ActionType_ACTION_TYPE_DELETE:
		return "deleted"
	default:
		panic(fmt.Errorf("unhandled action type '%s'", actionType))
	}
}

func mapToOrderByDirection(parsedDirection string) proto.OrderDirection {
	switch parsedDirection {
	case parser.OrderByAscending:
		return proto.OrderDirection_ORDER_DIRECTION_ASCENDING
	case parser.OrderByDescending:
		return proto.OrderDirection_ORDER_DIRECTION_DECENDING
	default:
		return proto.OrderDirection_ORDER_DIRECTION_UNKNOWN
	}
}

func (scm *Builder) applyJobAttribute(protoJob *proto.Job, attribute *parser.AttributeNode) {
	switch attribute.Name.Value {
	case parser.AttributePermission:
		protoJob.Permissions = append(protoJob.Permissions, scm.permissionAttributeToProtoPermission(attribute))
	case parser.AttributeSchedule:
		val, _, _ := resolve.ToValue[string](attribute.Arguments[0].Expression)
		src := strings.TrimPrefix(val, `"`)
		src = strings.TrimSuffix(src, `"`)

		c, _ := cron.Parse(src)

		protoJob.Schedule = &proto.Schedule{
			Expression: c.String(),
		}
	}
}

func (scm *Builder) applyJobInputs(protoMessage *proto.Message, inputs []*parser.JobInputNode) {
	for _, input := range inputs {
		protoField := &proto.MessageField{
			Name:        input.Name.Value,
			MessageName: protoMessage.GetName(),
			Type: &proto.TypeInfo{
				Type:     scm.parserTypeToProtoType(input.Type.Value),
				Repeated: input.Repeated,
			},
			Optional: input.Optional,
		}

		if protoField.GetType().GetType() == proto.Type_TYPE_ENUM {
			protoField.Type.EnumName = wrapperspb.String(input.Type.Value)
		}

		protoMessage.Fields = append(protoMessage.Fields, protoField)
	}
}

func (scm *Builder) applyFlowAttribute(protoFlow *proto.Flow, attribute *parser.AttributeNode) {
	switch attribute.Name.Value {
	case parser.AttributePermission:
		protoFlow.Permissions = append(protoFlow.Permissions, scm.permissionAttributeToProtoPermission(attribute))
	case parser.AttributeSchedule:
		val, _, _ := resolve.ToValue[string](attribute.Arguments[0].Expression)
		src := strings.Trim(val, `"`)

		c, _ := cron.Parse(src)

		protoFlow.Schedule = &proto.Schedule{
			Expression: c.String(),
		}
	}
}

func (scm *Builder) applyFlowInputs(protoMessage *proto.Message, inputs []*parser.FlowInputNode) {
	for _, input := range inputs {
		protoField := &proto.MessageField{
			Name:        input.Name.Value,
			MessageName: protoMessage.GetName(),
			Type: &proto.TypeInfo{
				Type:     scm.parserTypeToProtoType(input.Type.Value),
				Repeated: input.Repeated,
			},
			Optional: input.Optional,
		}

		if protoField.GetType().GetType() == proto.Type_TYPE_ENUM {
			protoField.Type.EnumName = wrapperspb.String(input.Type.Value)
		}

		if protoField.GetType().GetType() == proto.Type_TYPE_MODEL {
			protoField.Type.ModelName = wrapperspb.String(input.Type.Value)
		}

		protoMessage.Fields = append(protoMessage.Fields, protoField)
	}
}

// stripQuotes removes all double quotes from the given string, regardless of where they are.
func stripQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, "")
}

func makeInputMessageName(opName string, subMessageNames ...string) string {
	if len(subMessageNames) > 0 {
		subFieldNames := strings.Join(
			lo.Map(subMessageNames, func(s string, _ int) string { return casing.ToCamel(s) }),
			"")

		return fmt.Sprintf("%s%sInput", casing.ToCamel(opName), subFieldNames)
	}
	return fmt.Sprintf("%sInput", casing.ToCamel(opName))
}

func makeWhereMessageName(opName string) string {
	return fmt.Sprintf("%sWhere", casing.ToCamel(opName))
}

func makeOrderByMessageName(opName string, fieldName string) string {
	return fmt.Sprintf("%sOrderBy%s", casing.ToCamel(opName), casing.ToCamel(fieldName))
}

func makeValuesMessageName(opName string) string {
	return fmt.Sprintf("%sValues", casing.ToCamel(opName))
}

func makeMessageName(opName string) string {
	return fmt.Sprintf("%sMessage", casing.ToCamel(opName))
}

func makeSubscriberMessageName(subscriberName string) string {
	return fmt.Sprintf("%sEvent", casing.ToCamel(subscriberName))
}

func makeSubscriberMessageEventName(subscriberName string, modelName string, action string) string {
	return fmt.Sprintf("%s%s%sEvent", casing.ToCamel(subscriberName), casing.ToCamel(modelName), casing.ToCamel(action))
}

func makeSubscriberMessageEventTargetName(subscriberName string, modelName string, action string) string {
	return fmt.Sprintf("%s%s%sEventTarget", casing.ToCamel(subscriberName), casing.ToCamel(modelName), casing.ToCamel(action))
}

func makeEventName(modelName string, action string) string {
	return fmt.Sprintf("%s.%s", casing.ToSnake(modelName), action)
}

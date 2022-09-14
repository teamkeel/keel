package parser

// Keywords
const (
	KeywordModel      = "model"
	KeywordModels     = "models"
	KeywordApi        = "api"
	KeywordField      = "field"
	KeywordFields     = "fields"
	KeywordOperations = "operations"
	KeywordFunctions  = "functions"
	KeywordOperation  = "operation"
	KeywordFunction   = "function"
	KeywordDomains    = "domains"
	KeywordEmails     = "emails"
	KeywordRole       = "role"
	KeywordEnum       = "enum"
	KeywordWith       = "with"
)

// Built in Keel types. Worth noting a field type can also reference
// another user-defined model
const (
	FieldTypeID       = "ID"        // a uuid or similar
	FieldTypeText     = "Text"      // a string
	FieldTypeNumber   = "Number"    // an integer
	FieldTypeDate     = "Date"      // a date with no time element
	FieldTypeDatetime = "Timestamp" // a UTC unix timestamp
	FieldTypeBoolean  = "Boolean"   // a boolean
	FieldTypeSecret   = "Secret"    // an encrypted secret
	FieldTypePassword = "Password"  // a hashed password
)

var BuiltInTypes = map[string]bool{
	FieldTypeID:       true,
	FieldTypeText:     true,
	FieldTypeNumber:   true,
	FieldTypeDate:     true,
	FieldTypeDatetime: true,
	FieldTypeBoolean:  true,
	FieldTypeSecret:   true,
	FieldTypePassword: true,
}

func IsBuiltInFieldType(s string) bool {
	_, ok := BuiltInTypes[s]
	return ok
}

// Possible action types, applies to both "operations" and "functions"
const (
	ActionTypeGet    = "get"
	ActionTypeCreate = "create"
	ActionTypeUpdate = "update"
	ActionTypeList   = "list"
	ActionTypeDelete = "delete"
)

var ActionTypes = []string{
	ActionTypeGet,
	ActionTypeUpdate,
	ActionTypeCreate,
	ActionTypeList,
	ActionTypeDelete,
}

// All models get a field named "id" implicitly. This set of constants provides
// the set of this, and other similar implicit fields.
const (
	ImplicitFieldNameId        = "id"
	ImplicitFieldNameCreatedAt = "createdAt"
	ImplicitFieldNameUpdatedAt = "updatedAt"
)

const (
	ImplicitIdentityModelName         = "Identity"
	ImplicitIdentityFieldNameEmail    = "email"
	ImplicitIdentityFieldNamePassword = "password"
)

const (
	ImplicitAuthenticateOperationName = "authenticate"
)

const (
	AttributeUnique     = "unique"
	AttributePermission = "permission"
	AttributeWhere      = "where"
	AttributeSet        = "set"
	AttributeGraphQL    = "graphql"
	AttributePrimaryKey = "primaryKey"
	AttributeDefault    = "default"
	AttributeValidate   = "validate"
)

const (
	APITypeGraphQL = "graphql"
	APITypeRPC     = "rpc"
)

package parser

// Keywords
const (
	KeywordModel     = "model"
	KeywordApi       = "api"
	KeywordField     = "field"
	KeywordOperation = "operation"
	KeywordFunction  = "function"
)

// Built in Keel types. Worth noting a field type can also reference
// another user-defined model
const (
	FieldTypeID       = "ID"        // a uuid or similar
	FieldTypeIdentity = "Identity"  // represents the abstraction of a user/service account etc.
	FieldTypeText     = "Text"      // a string
	FieldTypeNumber   = "Number"    // an integer
	FieldTypeDate     = "Date"      // a date with no time element
	FieldTypeDatetime = "Timestamp" // a UTC unix timestamp
	FieldTypeBoolean  = "Boolean"   // a boolean
	FieldTypeImage    = "Image"     // an image file
	FieldTypeCurrency = "Currency"  // a currency value
	FieldTypeEnum     = "Enum"      // a field that can only contain a set of known values
)

// Possible action types, applies to both "operations" and "functions"
const (
	ActionTypeGet    = "get"
	ActionTypeCreate = "create"
	ActionTypeUpdate = "update"
	ActionTypeList   = "list"
	ActionTypeDelete = "delete"
)

// All models get a field named "id" implicitly. This set of constants provides
// the set of this, and other similar implicit fields.
const (
	ImplicitFieldNameId     = "id"
	ImplicitContextIdentity = "identity"
)

const (
	AttributeUnique     = "unique"
	AttributeOptional   = "optional"
	AttributePermission = "permission"
	AttributeWhere      = "where"
	AttributeSet        = "set"
	AttributeGraphQL    = "graphql"
	AttributePrimaryKey = "primaryKey"
)

const (
	APITypeGraphQL = "graphql"
	APITypeRPC     = "rpc"
)

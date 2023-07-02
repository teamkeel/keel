package parser

// Keywords
const (
	KeywordModel      = "model"
	KeywordModels     = "models"
	KeywordApi        = "api"
	KeywordMessage    = "message"
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
	KeywordReturns    = "returns"
)

// Types are roughly analogous to field types but they are used to type expressions
const (
	TypeNumber  = "Number"
	TypeText    = "Text"
	TypeBoolean = "Boolean"

	// These are unique to expressions
	TypeNull  = "Null"
	TypeArray = "Array"
	TypeIdent = "Ident"
	TypeEnum  = "Enum"
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

// Types for Message fields
const (
	MessageFieldTypeAny = "Any"
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

	// Arbitrary function action types
	ActionTypeRead  = "read"
	ActionTypeWrite = "write"
)

var OperationActionTypes = []string{
	ActionTypeGet,
	ActionTypeUpdate,
	ActionTypeCreate,
	ActionTypeList,
	ActionTypeDelete,
}

var FunctionActionTypes = []string{
	ActionTypeCreate,
	ActionTypeGet,
	ActionTypeDelete,
	ActionTypeList,
	ActionTypeUpdate,
	ActionTypeRead,
	ActionTypeWrite,
}

// All models get a field named "id" implicitly. This set of constants provides
// the set of this, and other similar implicit fields.
const (
	ImplicitFieldNameId        = "id"
	ImplicitFieldNameCreatedAt = "createdAt"
	ImplicitFieldNameUpdatedAt = "updatedAt"
)

var (
	ImplicitFieldNames = []string{ImplicitFieldNameId, ImplicitFieldNameCreatedAt, ImplicitFieldNameUpdatedAt}
)

const (
	ImplicitIdentityModelName           = "Identity"
	ImplicitIdentityFieldNameEmail      = "email"
	ImplicitIdentityFieldNamePassword   = "password"
	ImplicitIdentityFieldNameExternalId = "externalId"
	ImplicitIdentityFieldNameCreatedBy  = "createdBy"
)

const (
	AuthenticateOperationName         = "authenticate"
	RequestPasswordResetOperationName = "requestPasswordReset"
	PasswordResetOperationName        = "resetPassword"
)

const (
	AttributeUnique     = "unique"
	AttributePermission = "permission"
	AttributeWhere      = "where"
	AttributeSet        = "set"
	AttributePrimaryKey = "primaryKey"
	AttributeDefault    = "default"
	AttributeValidate   = "validate"
	AttributeRelation   = "relation"
	AttributeOrderBy    = "orderBy"
	AttributeSortable   = "sortable"
)

const (
	OrderByAscending  = "asc"
	OrderByDescending = "desc"
)

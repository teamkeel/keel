package parser

// Keywords
const (
	KeywordModel   = "model"
	KeywordModels  = "models"
	KeywordApi     = "api"
	KeywordMessage = "message"
	KeywordField   = "field"
	KeywordFields  = "fields"
	KeywordActions = "actions"
	KeywordDomains = "domains"
	KeywordEmails  = "emails"
	KeywordRole    = "role"
	KeywordEnum    = "enum"
	KeywordWith    = "with"
	KeywordReturns = "returns"
	KeywordJob     = "job"
	KeywordInput   = "inputs"
)

const (
	DefaultApi = "Api"
)

// Built in Keel types. Worth noting a field type can also reference
// another user-defined model
const (
	FieldTypeID        = "ID"        // a uuid or similar
	FieldTypeText      = "Text"      // a string
	FieldTypeNumber    = "Number"    // an integer
	FieldTypeDecimal   = "Decimal"   // a decimal
	FieldTypeDate      = "Date"      // a date with no time element
	FieldTypeTimestamp = "Timestamp" // a UTC unix timestamp
	FieldTypeBoolean   = "Boolean"   // a boolean
	FieldTypeSecret    = "Secret"    // an encrypted secret
	FieldTypePassword  = "Password"  // a hashed password
	FieldTypeMarkdown  = "Markdown"  // a markdown rich text
	FieldTypeVector    = "Vector"    // a vector
	FieldTypeFile      = "File"      // a inline file supplied as a data-url
)

// Types for Message fields
const (
	MessageFieldTypeAny = "Any"
)

var BuiltInTypes = map[string]bool{
	FieldTypeID:        true,
	FieldTypeText:      true,
	FieldTypeNumber:    true,
	FieldTypeDecimal:   true,
	FieldTypeDate:      true,
	FieldTypeTimestamp: true,
	FieldTypeBoolean:   true,
	FieldTypeSecret:    true,
	FieldTypePassword:  true,
	FieldTypeMarkdown:  true,
	FieldTypeVector:    true,
	FieldTypeFile:      true,
}

func IsBuiltInFieldType(s string) bool {
	_, ok := BuiltInTypes[s]
	return ok
}

// All possible action types
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

var ActionTypes = []string{
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
	FieldNameId        = "id"
	FieldNameCreatedAt = "createdAt"
	FieldNameUpdatedAt = "updatedAt"
)

var (
	FieldNames = []string{FieldNameId, FieldNameCreatedAt, FieldNameUpdatedAt}
)

const (
	IdentityModelName              = "Identity"
	IdentityFieldNameEmail         = "email"
	IdentityFieldNameEmailVerified = "emailVerified"
	IdentityFieldNamePassword      = "password"
	IdentityFieldNameExternalId    = "externalId"
	IdentityFieldNameIssuer        = "issuer"
	IdentityFieldNameName          = "name"
	IdentityFieldNameGivenName     = "givenName"
	IdentityFieldNameFamilyName    = "familyName"
	IdentityFieldNameMiddleName    = "middleName"
	IdentityFieldNameNickName      = "nickName"
	IdentityFieldNameProfile       = "profile"
	IdentityFieldNamePicture       = "picture"
	IdentityFieldNameWebsite       = "website"
	IdentityFieldNameGender        = "gender"
	IdentityFieldNameZoneInfo      = "zoneInfo"
	IdentityFieldNameLocale        = "locale"
)

const (
	RequestPasswordResetActionName = "requestPasswordReset"
	PasswordResetActionName        = "resetPassword"
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
	AttributeSchedule   = "schedule"
	AttributeFunction   = "function"
	AttributeOn         = "on"
	AttributeEmbed      = "embed"
	AttributeComputed   = "computed"
)

const (
	ThisVariable = "this"
)

const (
	OrderByAscending  = "asc"
	OrderByDescending = "desc"
)

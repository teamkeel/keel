package parser

// Built in Keel types. Worth noting a field type can also reference
// another user-defined model
const (
	FieldTypeID       = "ID"        // a uuid or similar
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

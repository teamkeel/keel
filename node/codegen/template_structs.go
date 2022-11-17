package codegenerator

// Template Structs defined here are used as friendly, presentational data structures
// that are passed to the various go templates as variables

// Represents a model in a .keel schema
type Model struct {
	Name   string
	Fields []*ModelField
}

type ModelField struct {
	Name string
	Type string
}

type Enum struct {
	Name   string
	Values []*EnumValue
}

type EnumValue struct {
	Label string
}

// Represents the database api to interact with each model defined
// in a Keel schema
type ModelApi struct {
	Name                string
	ModelName           string
	ModelNameLowerCamel string
	TableName           string
}

// Used to represent a custom function
// Initially will be used to generate friendly function wrappers that encapsulate
// the input and return types expected for custom functions
type CustomFunction struct {
	Name string
}

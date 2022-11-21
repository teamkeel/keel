package codegenerator

// Template Structs defined here are used as friendly, presentational data structures
// that are passed to the various go templates as variables

// Represents a model in a .keel schema
type Model struct {
	Name         string
	Fields       []*ModelField
	UniqueFields []*ModelField
}

type ModelField struct {
	Name           string
	IsOptional     bool
	Type           string // typescript type
	ConstraintType string // e.g StringConstraint / BooleanConstraint etc
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

type Action struct {
	Name          string
	OperationType OperationType // e.g Create / Update etc
	IsCustom      bool
	WriteInputs   []*ActionInput
	ReadInputs    []*ActionInput
	Inputs        []*ActionInput // includes inputs of Mode type Unknown (authenticate)
}

type ActionInput struct {
	Label          string
	Type           string
	IsOptional     bool
	Mode           InputMode
	ConstraintType string // e.g StringConstraint / BooleanConstraint etc
}

type OperationType string

const (
	OperationTypeCreate       OperationType = "Create"
	OperationTypeDelete       OperationType = "Delete"
	OperationTypeList         OperationType = "List"
	OperationTypeGet          OperationType = "Get"
	OperationTypeUpdate       OperationType = "Update"
	OperationTypeAuthenticate OperationType = "Authenticate"
)

type InputMode string

const (
	InputModeWrite   InputMode = "write"
	InputModeRead    InputMode = "read"
	InputModeUnknown InputMode = "unknown"
)

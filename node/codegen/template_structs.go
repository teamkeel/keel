package codegenerator

// Template Structs defined here are used as friendly, presentational data structures
// that are passed to the various go templates as variables (think "presenter" pattern).
// There is a need for these "simplified" data structures over passing structs from the
// proto.* package into templates, due to two reasons:
// 1. The API of proto structs isn't very friendly to templating (underlying enums with strange values for example)
// 2. Go's templating system is very basic - it has if / range statements but not much more, so there were
// some situations where it wasn't possible to enumerate underlying proto structures

// Represents a model in a .keel schema
type Model struct {
	Name           string
	NameLowerCamel string
	ApiName        string
	TableName      string
	Fields         []*ModelField
	UniqueFields   []*ModelField
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

type Action struct {
	Name          string
	OperationType OperationType // e.g Create / Update etc
	IsCustom      bool
	ModelName     string
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

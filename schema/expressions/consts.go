package expressions

import "github.com/teamkeel/keel/schema/parser"

type OperandPosition = string

const (
	OperandPositionLhs OperandPosition = "lhs"
	OperandPositionRhs OperandPosition = "rhs"
)

const (
	TypeStringMap = "StringMap"
)

// Defines which operators can be used for each field type
var operatorsForType = map[string][]string{
	parser.FieldTypeText: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
	parser.FieldTypeID: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
	parser.FieldTypeNumber: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorGreaterThan,
		parser.OperatorGreaterThanOrEqualTo,
		parser.OperatorLessThan,
		parser.OperatorLessThanOrEqualTo,
		parser.OperatorAssignment,
		parser.OperatorIncrement,
		parser.OperatorDecrement,
	},
	parser.FieldTypeBoolean: {
		parser.OperatorAssignment,
		parser.OperatorEquals,
		parser.OperatorNotEquals,
	},
	parser.FieldTypeDate: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorGreaterThan,
		parser.OperatorGreaterThanOrEqualTo,
		parser.OperatorLessThan,
		parser.OperatorLessThanOrEqualTo,
		parser.OperatorAssignment,
	},
	parser.FieldTypeDatetime: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorGreaterThan,
		parser.OperatorGreaterThanOrEqualTo,
		parser.OperatorLessThan,
		parser.OperatorLessThanOrEqualTo,
		parser.OperatorAssignment,
	},
	parser.TypeEnum: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
	parser.TypeArray: {
		parser.OperatorIn,
		parser.OperatorNotIn,
	},
	parser.TypeNull: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
	parser.TypeModel: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
	parser.FieldTypeMarkdown: {
		parser.OperatorEquals,
		parser.OperatorNotEquals,
		parser.OperatorAssignment,
	},
}

package actions

type Args struct {
	Values map[string]any
	Wheres map[string]any
}

// func (n *ArgsNormalizer) Normalize() (*ArgOutput, error) {
// 	args := n.m
// 	op := n.op

// 	o := &ArgOutput{}

// 	switch op.Type {
// 	case proto.OperationType_OPERATION_TYPE_LIST:
// 	case proto.OperationType_OPERATION_TYPE_UPDATE:
// 	case proto.OperationType_OPERATION_TYPE_CREATE:
// 	case proto.OperationType_OPERATION_TYPE_DELETE, proto.OperationType_OPERATION_TYPE_GET:
// 	default:
// 		panic("jhe")
// 	}

// 	return o, nil
// }

package options

// // RegisterAggregationFunctions adds custom aggregation functions that support predicates
// func RegisterAggregationFunctions() cel.EnvOption {
// 	return cel.Lib(aggregationLibrary{})
// }

// type aggregationLibrary struct{}

// func (aggregationLibrary) LibraryName() string {
// 	return "aggregation"
// }

// func (aggregationLibrary) CompileOptions() []cel.EnvOption {
// 	return []cel.EnvOption{
// 		cel.Function("sum",
// 			cel.MemberOverload("sum_list_predicate",
// 				[]*cel.Type{
// 					cel.ListType(cel.DynType),
// 					cel.StringType,
// 					cel.BoolType,
// 				},
// 				cel.DynType,
// 				cel.FunctionBinding(func(args ...ref.Val) ref.Val {
// 					if len(args) != 3 {
// 						return types.NewErr("invalid number of arguments")
// 					}
// 					list, ok := args[0].(ref.Val)
// 					if !ok {
// 						return types.NewErr("first argument must be a list")
// 					}
// 					_, ok = args[1].(types.String)
// 					if !ok {
// 						return types.NewErr("second argument must be a string")
// 					}
// 					predicate, ok := args[2].(types.Bool)
// 					if !ok {
// 						return types.NewErr("third argument must be a boolean")
// 					}

// 					iterator := list.(ref.Iterable).Iterator()
// 					var sum float64
// 					for iterator.HasNext() == types.True {
// 						item := iterator.Next()
// 						// Apply the predicate
// 						if predicate == types.True {
// 							if num, ok := item.(types.Number); ok {
// 								sum += num.Float64()
// 							}
// 						}
// 					}
// 					return types.Double(sum)
// 				}),
// 			),
// 		),
// 		cel.Function("filter",
// 			cel.MemberOverload("filter_list_predicate",
// 				[]*cel.Type{
// 					cel.ListType(cel.DynType),
// 					cel.StringType,
// 					cel.BoolType,
// 				},
// 				cel.ListType(cel.DynType),
// 				cel.FunctionBinding(func(args ...ref.Val) ref.Val {
// 					if len(args) != 3 {
// 						return types.NewErr("invalid number of arguments")
// 					}
// 					list, ok := args[0].(ref.Val)
// 					if !ok {
// 						return types.NewErr("first argument must be a list")
// 					}
// 					_, ok = args[1].(types.String)
// 					if !ok {
// 						return types.NewErr("second argument must be a string")
// 					}
// 					predicate, ok := args[2].(types.Bool)
// 					if !ok {
// 						return types.NewErr("third argument must be a boolean")
// 					}

// 					iterator := list.(ref.Iterable).Iterator()
// 					var filtered []ref.Val
// 					for iterator.HasNext() == types.True {
// 						item := iterator.Next()
// 						// Apply the predicate
// 						if predicate == types.True {
// 							filtered = append(filtered, item)
// 						}
// 					}
// 					return types.NewDynamicList(types.DefaultTypeAdapter, filtered)
// 				}),
// 			),
// 		),
// 	}
// }

// func (aggregationLibrary) ProgramOptions() []cel.ProgramOption {
// 	return []cel.ProgramOption{}
// }

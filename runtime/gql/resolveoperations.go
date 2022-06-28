package gql

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/teamkeel/keel/proto"
)

// A GetOperationResolver provides a Resolve method that matches the signature needed for
// a graphql.FieldResolveFn. It is implemented as an object that holds state, so that
// individual instances can be constructed with access to the data they need when their
// Resolve method gets called.
type GetOperationResolver struct {
	op *proto.Operation
}

func NewGetOperationResolver(op *proto.Operation) *GetOperationResolver {
	return &GetOperationResolver{
		op: op,
	}
}

func (r *GetOperationResolver) Resolve(p graphql.ResolveParams) (interface{}, error) {

	fmt.Printf("XXXX operation resolver fired\n")

	// For the moment - just illustrate how this resolver has the info it needs
	// to anticpate, which inputs are expected, and to fetch the corresponding
	// values from the incoming ResolveParams object.
	for _, input := range r.op.Inputs {
		paramValue, ok := p.Args[input.Name]
		if !ok {
			return nil, fmt.Errorf("the input named: %s is missing", input.Name)
		}
		fmt.Printf("XXXX found value: %v for param: %s\n", paramValue, input.Name)
	}

	// Todo. Code goes here to use the input paramaters, and knowledge of which type of
	// model we are "getting" to go off and find one/some.

	// Let's what happens if the resolver returns a map[string]any, that includes
	// (hard-coded) a field of the name our test is looking for - i.e. "name", and with
	// a string value.

	fmt.Printf("XXXX operation resolver returning arbitrary map object\n")
	res := map[string]any{
		"name":           "harriet",
		"someOtherField": "abcdef123",
	}

	return res, nil
}

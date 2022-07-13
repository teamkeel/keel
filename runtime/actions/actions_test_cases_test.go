package actions

var testCases []TestCase = []TestCase{
	{
		KeelSchema: `
			model Person {
				fields {
					name Text @unique
				}
				operations {
					create createPerson() with (name)
				}
			}
			`,
		ModelName:     "Person",
		OperationName: "createPerson",
		OperationInputs: map[string]any{
			"foo": 42,
		},
		ExpectedActionResponse: "wontbethis",
		ExpectedSQLResponse:    "wontbethis",
	},
}

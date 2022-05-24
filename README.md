# keel-cli

keel-cli is a tool to build and deploy services

## Usage

| Command           | Description                                                                                       |
|-------------------|---------------------------------------------------------------------------------------------------|
|  build            | Build the application ready for production deployment                                             |
|  completion       | Generate the autocompletion script for the specified shell                                        |
|  diff             | Read DB migrations directory, construct the schema and diff the two                               |
|  help             | Help about any command                                                                            |
|  run              | Run the application locally                                                                       |
|  validate         | Validate the Keel schema                                                                          |


## Development

Requires Go 1.18.x

## Testing

### Generating new schema test cases

There is a handy helper script to generate the relevant test case files if you'd like to write a new test case for something in the schema / validation rules:

```
sh ./scripts/new-test-case test_case_name
```

This will generate:

- A blank `schema.keel` file
- An `errors.json` file where you can assert what errors you are expecting

### A note on test cases

Naming convention for test cases should describe the units you are testing in a hierarchical manner. Test cases covering validations should begin with `validation_` for example, and subsequent fragments should describe which validation is being tested. 

If a test case is complex, or the name itself doesn't adequately describe what it is testing, then either reconsider the naming, or consider adding a `description.md` file into the directory where the test case lives to provide more colour.
m

### Running individual test cases

You can run a test by it's pattern like so:

```
go test -timeout 30s -run ^TestSchema/test_case_name github.com/teamkeel/keel/schema
```
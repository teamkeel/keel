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

## Setting up

Run:

```
sh ./scripts/setup.sh
```

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


### Running individual test cases

You can run a test by it's pattern like so:

```
go test -timeout 30s -run ^TestSchema/test_case_name github.com/teamkeel/keel/schema
```

### Run all test cases

You can run all of the test cases contained with `schema/testdata` by running:

```
./scripts/run-all-test-cases.sh
```

## Contributing

This repo is setup so that all contributors follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) conventions when writing commit messages.

__You will not be able to merge your PR without following these conventions.__

### A short primer on conventional commit messages

Prefix your commit messages with:

- For minor version bumps: `feat: some commit message`
- For patches / fixes: `fix: some commit message`
- For major bumps (note the `!`): `fix!: a major / breaking change`

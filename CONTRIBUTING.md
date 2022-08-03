# Contributing

This repo is setup so that all contributors follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) conventions when writing commit messages.

__You will not be able to merge your PR without following these conventions.__

## A short primer on conventional commit messages

Prefix your commit messages with:

- For minor version bumps: `feat: some commit message`
- For patches / fixes: `fix: some commit message`
- For major bumps (note the `!`): `fix!: a major / breaking change`

## Testing

### Generating new schema test cases

There is a handy helper script to generate the relevant test case files if you'd like to write a new test case for something in the schema / validation rules:

```
sh ./scripts/new-test-case test_case_name
```

This will generate:

- A blank `schema.keel` file
- An `errors.json` file where you can assert what errors you are expecting

### Updating test cases

If you have made a change to error message copy, then you can run `make testdata` and all of the test case JSON files will be updated to the latest copy.

### A note on test cases

Naming convention for test cases should describe the units you are testing in a hierarchical manner. Test cases covering validations should begin with `validation_` for example, and subsequent fragments should describe which validation is being tested.

If a test case is complex, or the name itself doesn't adequately describe what it is testing, then either reconsider the naming, or consider adding a `description.md` file into the directory where the test case lives to provide more colour.

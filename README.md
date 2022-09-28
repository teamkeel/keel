# keel-cli

keel-cli is a tool to build and deploy services

## Usage

| Command    | Description                                                         |
| ---------- | ------------------------------------------------------------------- |
| build      | Build the application ready for production deployment               |
| completion | Generate the autocompletion script for the specified shell          |
| diff       | Read DB migrations directory, construct the schema and diff the two |
| help       | Help about any command                                              |
| run        | Run the application locally                                         |
| validate   | Validate the Keel schema                                            |
| generate   | Generates requisite types and runtime code for custom functions     |

## Testing

The [`@teamkeel/testing`](https://github.com/teamkeel/keel/tree/main/testing/package) package allows us to write end-to-end tests for the Keel runtime.

Our internal tests are located in the `integration/testdata` directory. The integration test framework is a "meta test", meaning that in addition to any assertions we make in our `*.test.ts`, there is a corresponding `expected.json` file that documents which test cases should pass or fail. This is especially useful if we have test cases we intentionally want to fail.

Each directory is its own isolated application. At the very least, a test case directory should have:

- A `schema.keel` file
- At least one test file, following the glob pattern `*.test.ts`
- An `expected.json` file, which documents which test cases should pass or fail

If you are testing custom functions, then you should include a `functions/` sub directory, with custom function files inside.

### Expectations and assertions

The testing framework allows you to setup data in the database, call both built-in and custom actions, as well as assert on the return value of actions, fulfilling the AAA (arrange-act-assert) testing strategy. The database is cleared in between each test *file*, but isn't yet cleared between individual test cases within a file (see 'Gotchas' section below):

```typescript
import { test, expect, Actions, MyModel } from '@teamkeel/testing'

test('a sample test', async () => {
  // arrange - setup the data
  const { object: createdModel } = await MyModel.create({ foo: 'bar' });

  // you can also fetch things from the database:
  const { collection } = await MyModel.findMany({ foo: 'bar' });
  const { object } = await MyModel.findOne({ myUniqueField: 'xxx' });

  // act - call an action
  const { object: result } = await Actions.getMyModel({ id: createdModel.id });

  // assert on the result
  expect.equal(result.id, createdModel.id);
})
```

### Running all of the integration tests

You can simply run:

```
make testintegration
```

### Running specific integration tests

#### Via the command line

If you want to run all TypeScript tests in a particular test case directory, you can run:

```
go test ./integration -run ^TestIntegration/my_test_case_dir_name -count=1 -v
```

However, this doesn't allow you to isolate an individual test case written in a `test()` block in a `*.test.ts` file. In order to do this, you must pass an additional `pattern` flag to the test command like so:

```
go test ./integration -run ^TestIntegration/built_in_actions -count=1 -v -pattern "list action"
```

The `pattern` can either be a simple string matching a test case name, or you can provide a regex:

```
go test ./integration -run ^TestIntegration/built_in_actions -count=1 -v -pattern "get action (.*)"
```

The above will run all tests that begin with "get action".

#### Running / debugging tests via vscode test runner

You can use the `Run / Debug integration test` launch configuration in VSCode. When you click the "Play" button, it will ask you to input the test case directory (relative from the root `integration/testdata`) as well as to input any test name patterns which allows you to either isolate on an individual JavaScript/TypeScript test or you can provide a Regular Expression.

This allows you to debug individual tests using the VSCode debugger, although you will not be able to debug the TypeScript test code at the moment - it is possible in the future we may be able to do this as we can pass additional flags to enable debugging to `ts-node` when running a test file via the test harness.

### Gotchas

- The database isn't yet cleared between individual `test()` blocks - this will be coming soon.
- The test runner actually copies each test case directory to a temporary directory elsewhere when executing the test - inside of the temporary directory, all of the dependencies necessary to have a viable application are installed (this also includes creation of package.json and tsconfig.json). This saves us having to populate each directory each time we add a new test case. The downside is that we do not get intellisense when writing our tests; this would be tricky anyway given a lot of the code inside of the `@teamkeel/testing` package is code generated at the point of running the test. 

## Development

You need the following installed:

- Go `brew install go`
- Node - first install [`pnpm`](https://pnpm.io/installation) then run `pnpm env use --global lts`
- Docker - https://docs.docker.com/desktop/install/mac-install/
- libpq - `brew install libpq` and follow post-install Brew instructions on updating PATH

A working setup will look something like this (paths will vary):

```sh
~/code/keel main $ which go
/usr/local/go/bin/go
~/code/keel main $ which node
/Users/jonbretman/.nvm/versions/node/v16.16.0/bin/node
~/code/keel main $ which docker
/usr/local/bin/docker
~/code/keel main $ which psql
/opt/homebrew/opt/libpq/bin/psql
```

### Setting up conventional commits

Run the following setup command:

```
sh ./scripts/setup.sh
```

### Using the CLI in development

```
go run cmd/keel/main.go [cmd] [args]
```

## Building from source

You can build the CLI executable by running:

```
make
```

And to interact with the executable version of the CLI, simply run:

```
./keel validate -f ...
```

# Contributing

Please read the [Contribution guidelines](/CONTRIBUTING.md) before lodging a PR.

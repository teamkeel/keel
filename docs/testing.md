# Testing

We provide the ability to test the internals of the Keel runtime through our own custom testing framework, which allows us to write tests against any given schema. We can:

- Work with strongly typed TypeScript types that are directly based on the constructs in our schema
- Execute schema actions and get the response back from each action and assert on the result.

We have a large array of integration tests that exist inside of the `integration/testdata` directory. Each directory is a standalone Keel application. For each test case directory, bootstrapping of build files such as `package.json`, `tsconfig.json`, `node_modules` etc is automatically handled for you by the `node` package.

## Writing Tests

The test framework will search for all files matching the pattern `*.test.ts` inside a given directory.

The testing framework utilizes [vitest](https://vitest.dev/) as its test runner on the JavaScript side.

We have a custom JavaScript package `@teamkeel/testing` that contains dynamically generated types and code based on your schema. This package is updated in real time with any changes to a schema when you run a test via any of the methods described in the `Executing Tests` section below.

An example test might look like:

```typescript
import { actions, resetDatabase, models } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("testing an action works" async () => {
  const post = await actions.create({
    title: '123'
  });

  expect(post.title).toEqual('123');
});

test("testing the db api", async () => {
  const post = await models.post.create({
    title: '123'
  });

  expect(post.title).toEqual('123');
});
```

## Executing Tests

There are three ways of executing tests - via a special Go integration test, and via the exposed `keel test` command. If you'd prefer to see the vitest output (including the summary of tests executed) in full, using the `keel test` command is preferable.

### Via the Keel test cmd

You can run individual test cases via:

```
go run cmd/keel/main.go test -d ./integration/testdata/your_test_case
```

You can also isolate individual JavaScript `test()` blocks within each test case by using the `--pattern` flag:

```
go run cmd/keel/main.go test -d ./integration/testdata/with_custom_function --pattern "fetching(.*)"
```

### Via `go test`

You can run the whole of the integration test suite by running:

```
make testintegration
```

Or run a single test case:

```
 go test ./integration -run ^TestIntegration/with_custom_function -v
```

### Via the VSCode `launch.json` configurations

Several debug configurations are provided in the `.vscode/launch.json`. The debugger only works for `.go` code that orchestrates each test - unfortunately it isn't possible (yet) to debug the actual JavaScript code in each `*.test.ts` file.

#### Debug the `test` cmd

- Go to the `Run and Debug` pane in VSCode, select the `Debug test` configuration, and press the 'Play' button
- Enter the path to the directory containing a `.keel` schema file.
- Press enter
- Enter a regular expression pattern. Defaults to `(.*)` (all tests)
- Press enter, and debug!

# `@teamkeel/testing`

\*\* All examples use the following sample schema:

```
model Post {
  fields {
    title Text
  }

  operations {
    create createPost() with(title) {
      @permission(post.title != "bar")
    }
  }
}

api Web {
  @graphql

  models {
    Post
  }
}
```

## Overview

The [`@teamkeel/testing`](https://github.com/teamkeel/keel/tree/main/testing/package) package allows us to write end-to-end tests for the Keel runtime.

Our internal tests are located in the `integration/testdata` directory. The integration test framework is a "meta test", meaning that in addition to any assertions we make in our `*.test.ts`, there is a corresponding `expected.json` file that documents which test cases should pass or fail. This is especially useful if we have test cases we intentionally want to fail.

Each directory is its own isolated application. At the very least, a test case directory should have:

- A `schema.keel` file
- At least one test file, following the glob pattern `*.test.ts`
- An `expected.json` file, which documents which test cases should pass or fail

If you are testing custom functions, then you should include a `functions/` sub directory, with custom function files inside.

Define new test cases using:

```typescript
import { test } from "@teamkeel/testing";

test("it does something", async () => {});
```

The testing framework allows you to setup data in the database, call both built-in and custom actions, as well as assert on the return value of actions, fulfilling the AAA (arrange-act-assert) testing strategy:

```typescript
import { test, expect, Actions, MyModel } from "@teamkeel/testing";

test("a sample test", async () => {
  // arrange - setup the data
  const { object: createdModel } = await MyModel.create({ foo: "bar" });

  // you can also fetch things from the database:
  const { collection } = await MyModel.findMany({ foo: "bar" });
  const { object } = await MyModel.findOne({ myUniqueField: "xxx" });

  // act - call an action
  const { object: result } = await Actions.getMyModel({ id: createdModel.id });

  // assert on the result
  expect(result.id).toEqual(createdModel.id);
});
```

## Querying the database

You can create / update and query data easily using our Database API. Complete Database API documentation can be found at XXX

```typescript
import { test, Post } from "@teamkeel/testing";

test("it does something", async () => {
  const { object: createdPost } = await Post.create({
    title: "a title",
  });

  expect(createdPost.title).toEqual("a title");
});
```

## Calling actions

You can execute Keel actions like so:

```typescript
import { test, actions } from "@teamkeel/testing";

test("it does something", async () => {
  const { object: createdPost } = await actions.createPost({
    title: "a title",
  });

  expect(createdPost.title).toEqual("a title");
});
```

## Expectation API

The table below outlines our expectation API:

| Expectation                   | Actual type    | Expected type | Example                                                                                        |
| ----------------------------- | -------------- | ------------- | ---------------------------------------------------------------------------------------------- |
| `toEqual`                     | `any`          | `any`         | `expect(1).toEqual(1)`                                                                         |
| `notToEqual`                  | `any`          | `any`         | `expect(1).notToEqual(2)`                                                                      |
| `toHaveError`                 | `ActionResult` | `ActionError` | `expect(await actions.createPost({ title: 'too long' })).toHaveError({ message: 'too long' })` |
| `notToHaveError`              | `ActionResult` | `ActionError` | `expect(await actions.createPost({ title: 'OK' })).notToHaveError()`                           |
| `toHaveAuthorizationError`    | `ActionResult` | N/A           | `expect(await actions.createPost({ title: 'bar' })).toHaveAuthorizationError()`                |
| `notToHaveAuthorizationError` | `ActionResult` | N/A           | `expect(await actions.createPost({ title: 'foo' })).notToHaveAuthorizationError()`             |
| `toBeEmpty`                   | `any`          | N/A           | `expect(null).toBeEmpty()`                                                                     |
| `notToBeEmpty`                | `any`          | N/A           | `expect({ foo: 'bar' }).notToBeEmpty()`                                                        |
| `toContain`                   | `Array<T>`     | `any`         | `expect([{ foo: 'bar' }]).toContain({ foo: 'bar' })`                                           |
| `notToContain`                | `Array<T>`     | `any`         | `expect([]).notToContain({ foo: 'bar' })`                                                      |

## Running integration tests

To run all of the integration tests, you can simply run:

```
make testintegration
```

## Running specific integration tests

### Via the command line

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

### Running / debugging tests via vscode test runner

You can use the `Run / Debug integration test` launch configuration in VSCode. When you click the "Play" button, it will ask you to input the test case directory (relative from the root `integration/testdata`) as well as to input any test name patterns which allows you to either isolate on an individual JavaScript/TypeScript test or you can provide a Regular Expression.

This allows you to debug individual tests using the VSCode debugger, although you will not be able to debug the TypeScript test code at the moment - it is possible in the future we may be able to do this as we can pass additional flags to enable debugging to `ts-node` when running a test file via the test harness.

### Gotchas

- The test runner actually copies each test case directory to a temporary directory elsewhere when executing the test - inside of the temporary directory, all of the dependencies necessary to have a viable application are installed (this also includes creation of package.json and tsconfig.json). This saves us having to populate each directory each time we add a new test case. The downside is that we do not get intellisense when writing our tests; this would be tricky anyway given a lot of the code inside of the `@teamkeel/testing` package is code generated at the point of running the test.

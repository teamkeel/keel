# `@teamkeel/testing`

** All examples use the following sample schema:

```
model Post {
  fields {
    title Text
  }

  operations {
    create createPost() with(title)
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

Define new test cases using:

```typescript
import { test } from '@teamkeel/testing';

test('it does something', async () => {

})
```

## Querying the database

You can create / update and query data easily using our Database API. The full post API can be found at XXX

```typescript
import { test, Post } from '@teamkeel/testing';

test('it does something', async () => {
  const { object: createdPost } = await Post.create({
    title: 'a title'
  });

  expect(createdPost.title).toEqual('a title');
})
```

## Calling actions

You can execute Keel actions like so:

```typescript
import { test, actions } from '@teamkeel/testing';

test('it does something', async () => {
  const { object: createdPost } = await actions.createPost({
    title: 'a title'
  });

  expect(createdPost.title).toEqual('a title');
})
```

## Assertions + expectations

The table below outlines our expectation API:

| Expectation                   | Actual type    | Expected type | Example                                                                                        |
|-------------------------------|----------------|---------------|------------------------------------------------------------------------------------------------|
| `toEqual`                     | `any`          | `any`         | `expect(1).toEqual(1)`                                                                         |
| `notToEqual`                  | `any`          | `any`         | `expect(1).notToEqual(2)`                                                                      |
| `toHaveError`                 | `ActionResult` | `ActionError` | `expect(await actions.createPost({ title: 'too long' })).toHaveError({ message: 'too long' })` |
| `notToHaveError`              | `ActionResult` | `ActionError` | `expect(await actions.createPost({ title: 'OK' })).notToHaveError()`                           |
| `toHaveAuthorizationError`    | `ActionResult` | N/A           | `expect(await actions.forbiddenAction({ foo: 'bar' })).toHaveAuthorizationError()`             |
| `notToHaveAuthorizationError` | `ActionResult` | N/A           | `expect(await actions.permittedAction({ foo: 'bar' })).notToHaveAuthorizationError()`          |
| `toBeEmpty`                   | `any`          | N/A           | `expect(null).toBeEmpty()`                                                                     |
| `notToBeEmpty`                | `any`          | N/A           | `expect({ foo: 'bar' }).notToBeEmpty()`                                                        |
| `toContain`                   | `Array<T>`     | `any`         | `expect([{ foo: 'bar' }]).toContain({ foo: 'bar' })`                                           |
| `notToContain`                | `Array<T>`     | `any`         | `expect([]).notToContain({ foo: 'bar' })`                                                      |
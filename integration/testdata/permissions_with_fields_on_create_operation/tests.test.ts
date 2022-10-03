import { test, expect, actions, Post } from '@teamkeel/testing'

test('text permission successful', async () => {
  expect(
    await actions
    .createPostTextPermission({ title: "hello" })
  ).notToHaveAuthorizationError()
})

test('text permission failed', async () => {
  expect(
    await actions
    .createPostTextPermission({ title: "not hello" })
  ).toHaveAuthorizationError(true)
})

test('number permission successful', async () => {
  expect(
    await actions
    .createPostNumberPermission({ views: 5 })
  ).notToHaveAuthorizationError()
})

test('number permission failed', async () => {
  expect(
    await actions
    .createPostNumberPermission({ views: 500 })
  ).toHaveAuthorizationError()
})

test('boolean permission successful', async () => {
  expect(
    await actions
    .createPostBooleanPermission({ active: true })
  ).notToHaveAuthorizationError()
})

test('boolean permission failed', async () => {
  expect(
    await actions
    .createPostBooleanPermission({ active: false })
  ).toHaveAuthorizationError()
})

test('text not null permission successful', async () => {
  expect(
    await actions
    .createPostTextNotNullPermission({ title: "hello" })
  ).notToHaveAuthorizationError()
})

test('text not null permission failed', async () => {
  expect(
    await actions
    .createPostTextNotNullPermission({ title: null })
  ).toHaveAuthorizationError()
})

test('text null permission failed', async () => {
  expect(
    await actions
    .createPostTextNullPermission({})
  ).toHaveAuthorizationError()
})

test('number not null permission successful', async () => {
  expect(
    await actions
    .createPostNumberNotNullPermission({ views: 5 })
  ).notToHaveAuthorizationError()
})

test('number not null permission failed', async () => {
  expect(
    await actions
    .createPostNumberNotNullPermission({ views: null })
  ).toHaveAuthorizationError()
})

test('number null permission failed', async () => {
  expect(
    await actions
    .createPostNumberNullPermission({})
  ).toHaveAuthorizationError()
})

test('boolean not null permission successful', async () => {
  expect(
    await actions
    .createPostBooleanNotNullPermission({ active: true })
  ).notToHaveAuthorizationError()
})

test('boolean not null permission failed', async () => {
  expect(
    await actions
    .createPostBooleanNotNullPermission({ active: null })
  ).toHaveAuthorizationError()
})

test('boolean null permission failed', async () => {
  expect(
    await actions
    .createPostBooleanNullPermission({})
  ).toHaveAuthorizationError()
})

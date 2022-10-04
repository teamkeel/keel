import { test, expect, actions, Post } from '@teamkeel/testing'

test('text permission successful', async () => {
  const { object: post } = await actions
    .createPostWithTitle({ title: "hello" })

  expect(
    await actions
    .deletePostTextPermission({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('text permission failed', async () => {
  const { object: post } = await actions
    .createPostWithTitle({ title: "not hello" })

  expect(
    await actions
    .deletePostTextPermission({ id: post.id })
  ).toHaveAuthorizationError()
})

test('number permission successful', async () => {
  const { object: post } = await actions
    .createPostWithViews({ views: 5 })

  expect(
    await actions
    .deletePostNumberPermission({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('number permission failed', async () => {
  const { object: post } = await actions
    .createPostWithViews({ views: 500 })

  expect(
    await actions
    .deletePostNumberPermission({ id: post.id })
  ).toHaveAuthorizationError()
})

test('boolean permission successful', async () => {
  const { object: post } = await actions
    .createPostWithActive({ active: true })

  expect(
    await actions
    .deletePostBooleanPermission({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('boolean permission failed', async () => {
  const { object: post } = await actions
    .createPostWithActive({ active: false })

  expect(
    await actions
    .deletePostBooleanPermission({ id: post.id })
  ).toHaveAuthorizationError()
})

test('text not null permission successful', async () => {
  const { object: post } = await actions
    .createPostWithTitle({ title: "hello" })

  expect(
    await actions
    .deletePostTextNotNullPermission({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('text not null permission failed', async () => {
  const { object: post } = await actions
    .createPostWithTitle({ title: null })

  expect(
    await actions
    .deletePostTextNotNullPermission({ id: post.id })
  ).toHaveAuthorizationError()
})

test('text not null from default permission failed', async () => {
  const { object: post } = await actions
    .createPost({})

  expect(
    await actions
    .deletePostTextNotNullPermission({ id: post.id })
  ).toHaveAuthorizationError()
})

test('number not null permission successful', async () => {
  const { object: post } = await actions
    .createPostWithViews({ views: 5 })

  expect(
    await actions
    .deletePostNumberNotNullPermission({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('number not null permission failed', async () => {
  const { object: post } = await actions
    .createPostWithViews({ views: null })

  expect(
    await actions
    .deletePostNumberNotNullPermission({ id: post.id })
  ).toHaveAuthorizationError()
})

test('number not null from default permission failed', async () => {
  const { object: post } = await actions
    .createPost({})

  expect(
    await actions
    .deletePostNumberNotNullPermission({ id: post.id })
  ).toHaveAuthorizationError()
})

test('boolean not null permission successful', async () => {
  const { object: post } = await actions
    .createPostWithActive({ active: true })

  expect(
    await actions
    .deletePostBooleanNotNullPermission({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('boolean not null permission failed', async () => {
  const { object: post } = await actions
    .createPostWithActive({ active: null })

  expect(
    await actions
    .deletePostBooleanNotNullPermission({ id: post.id })
  ).toHaveAuthorizationError()
})

test('boolean not null from default permission failed', async () => {
  const { object: post } = await actions
    .createPost({})

  expect(
    await actions
    .deletePostBooleanNotNullPermission({ id: post.id })
  ).toHaveAuthorizationError()
})

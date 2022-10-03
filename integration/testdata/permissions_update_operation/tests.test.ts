import { test, expect, actions, Post } from '@teamkeel/testing'

test('text literal permission successful', async () => {
  const { object: post } = await actions
    .createWithText({ title: "hello" })

  expect(
    await actions
      .updateTextPermission({ 
        where: { id: post.id },
        values: { title: "goodbye" }
      })
  ).notToHaveAuthorizationError()
})

test('text literal permission failed', async () => {
  const { object: post } = await actions
    .createWithText({ title: "goodbye" })

  expect(
    await actions
      .updateTextPermission({ 
        where: { id: post.id },
        values: { title: "hello" }
      })
  ).toHaveAuthorizationError()
})

test('text null literal permission failed', async () => {
  const { object: post } = await actions
    .createWithText({ title: "goodbye" })

  expect(
    await actions
      .updateTextPermission({ 
        where: { id: post.id },
        values: { title: null }
      })
  ).toHaveAuthorizationError()
})

test('number literal permission successful', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 1 })

  expect(
    await actions
      .updateNumberPermission({ 
        where: { id: post.id },
        values: { views: 100 } 
      })
  ).notToHaveAuthorizationError()
})

test('number literal permission failed', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 100 })

  expect(
    await actions
      .updateNumberPermission({ 
        where: { id: post.id },
        values: { views: 1 }
      })
  ).toHaveAuthorizationError()
})

test('number null literal permission failed', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 100 })

  expect(
    await actions
      .updateNumberPermission({ 
        where: { id: post.id },
        values: { views: null }
      })
  ).toHaveAuthorizationError()
})

test('boolean literal permission successful', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: true })

  expect(
    await actions
      .updateBooleanPermission({ 
        where: { id: post.id },
        values: { active: false }
      })
  ).notToHaveAuthorizationError()
})

test('boolean literal permission failed', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: false })

  expect(
    await actions
      .updateBooleanPermission({ 
        where: { id: post.id },
        values: { active: true }
      })
  ).toHaveAuthorizationError()
})

test('boolean null literal permission failed', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: false })

  expect(
    await actions
      .updateBooleanPermission({ 
        where: { id: post.id },
        values: { active: null }
      })
  ).toHaveAuthorizationError()
})

test('text field permission successful', async () => {
  const { object: post } = await actions
    .createWithText({ title: "hello" })

  expect(
    await actions
      .updateFromFieldTextPermission({ 
        where: { id: post.id },
        values: { title: "goodbye" }
      })
  ).notToHaveAuthorizationError()
})

test('text field permission failed', async () => {
  const { object: post } = await actions
    .createWithText({ title: "goodbye" })

  expect(
    await actions
      .updateFromFieldTextPermission({ 
        where: { id: post.id },
        values: { title: "hello" }
      })
  ).toHaveAuthorizationError()
})

test('text null field permission failed', async () => {
  const { object: post } = await actions
    .createWithText({ title: "goodbye" })

  expect(
    await actions
      .updateFromFieldTextPermission({ 
        where: { id: post.id },
        values: { title: null }
      })
  ).toHaveAuthorizationError()
})

test('number field permission successful', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 1 })

  expect(
    await actions
      .updateFromFieldNumberPermission({ 
        where: { id: post.id },
        values: { views: 100 } 
      })
  ).notToHaveAuthorizationError()
})

test('number field permission failed', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 100 })

  expect(
    await actions
      .updateFromFieldNumberPermission({ 
        where: { id: post.id },
        values: { views: 1 }
      })
  ).toHaveAuthorizationError()
})

test('number null field permission failed', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 100 })

  expect(
    await actions
      .updateFromFieldNumberPermission({ 
        where: { id: post.id },
        values: { views: null }
      })
  ).toHaveAuthorizationError()
})

test('boolean field permission successful', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: true })

  expect(
    await actions
      .updateFromFieldBooleanPermission({ 
        where: { id: post.id },
        values: { active: false }
      })
  ).notToHaveAuthorizationError()
})

test('boolean field permission failed', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: false })

  expect(
    await actions
      .updateFromFieldBooleanPermission({ 
        where: { id: post.id },
        values: { active: true }
      })
  ).toHaveAuthorizationError()
})

test('boolean null field permission failed', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: false })

  expect(
    await actions
      .updateFromFieldBooleanPermission({ 
        where: { id: post.id },
        values: { active: null }
      })
  ).toHaveAuthorizationError()
})

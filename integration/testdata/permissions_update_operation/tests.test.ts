import { test, expect, actions, Post, Identity } from '@teamkeel/testing'

test('string permission with literal is authorized', async () => {
  const { object: post } = await actions
    .createWithText({ title: "hello" })

  expect(
    await actions
      .updateWithTextPermissionLiteral({ 
        where: { id: post.id },
        values: { title: "goodbye" }
      })
  ).notToHaveAuthorizationError()
})

test('string permission with literal is not authorized', async () => {
  const { object: post } = await actions
    .createWithText({ title: "goodbye" })

  expect(
    await actions
      .updateWithTextPermissionLiteral({ 
        where: { id: post.id },
        values: { title: "hello" }
      })
  ).toHaveAuthorizationError()
})

test('string permission with null literal is not authorized', async () => {
  const { object: post } = await actions
    .createWithText({ title: "goodbye" })

  expect(
    await actions
      .updateWithTextPermissionLiteral({ 
        where: { id: post.id },
        values: { title: null }
      })
  ).toHaveAuthorizationError()
})

test('number permission with literal is authorized', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 1 })

  expect(
    await actions
      .updateWithNumberPermissionLiteral({ 
        where: { id: post.id },
        values: { views: 100 } 
      })
  ).notToHaveAuthorizationError()
})

test('number permission with literal is not authorized', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 100 })

  expect(
    await actions
      .updateWithNumberPermissionLiteral({ 
        where: { id: post.id },
        values: { views: 1 }
      })
  ).toHaveAuthorizationError()
})

test('number permission with null literal is not authorized', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 100 })

  expect(
    await actions
      .updateWithNumberPermissionLiteral({ 
        where: { id: post.id },
        values: { views: null }
      })
  ).toHaveAuthorizationError()
})

test('boolean permission with literal is authorized', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: true })

  expect(
    await actions
      .updateWithBooleanPermissionLiteral({ 
        where: { id: post.id },
        values: { active: false }
      })
  ).notToHaveAuthorizationError()
})

test('boolean permission with literal is not authorized', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: false })

  expect(
    await actions
      .updateWithBooleanPermissionLiteral({ 
        where: { id: post.id },
        values: { active: true }
      })
  ).toHaveAuthorizationError()
})

test('boolean permission with null literal is not authorized', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: false })

  expect(
    await actions
      .updateWithBooleanPermissionLiteral({ 
        where: { id: post.id },
        values: { active: null }
      })
  ).toHaveAuthorizationError()
})

test('string permission with field succeeds', async () => {
  const { object: post } = await actions
    .createWithText({ title: "hello" })

  expect(
    await actions
      .updateWithTextPermissionFromField({ 
        where: { id: post.id },
        values: { title: "goodbye" }
      })
  ).notToHaveAuthorizationError()
})

test('string permission with field is not authorized', async () => {
  const { object: post } = await actions
    .createWithText({ title: "goodbye" })

  expect(
    await actions
      .updateWithTextPermissionFromField({ 
        where: { id: post.id },
        values: { title: "hello" }
      })
  ).toHaveAuthorizationError()
})

test('string permission with null field is not authorized', async () => {
  const { object: post } = await actions
    .createWithText({ title: "goodbye" })

  expect(
    await actions
      .updateWithTextPermissionFromField({ 
        where: { id: post.id },
        values: { title: null }
      })
  ).toHaveAuthorizationError()
})

test('number permission with field is authorized', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 1 })

  expect(
    await actions
      .updateWithNumberPermissionFromField({ 
        where: { id: post.id },
        values: { views: 100 } 
      })
  ).notToHaveAuthorizationError()
})

test('number permission with field is not authorized', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 100 })

  expect(
    await actions
      .updateWithNumberPermissionFromField({ 
        where: { id: post.id },
        values: { views: 1 }
      })
  ).toHaveAuthorizationError()
})

test('number permission with null field is not authorized', async () => {
  const { object: post } = await actions
    .createWithNumber({ views: 100 })

  expect(
    await actions
      .updateWithNumberPermissionFromField({ 
        where: { id: post.id },
        values: { views: null }
      })
  ).toHaveAuthorizationError()
})

test('boolean permission with field is authorized', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: true })

  expect(
    await actions
      .updateWithBooleanPermissionFromField({ 
        where: { id: post.id },
        values: { active: false }
      })
  ).notToHaveAuthorizationError()
})

test('boolean permission with field is not authorized', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: false })

  expect(
    await actions
      .updateWithBooleanPermissionFromField({ 
        where: { id: post.id },
        values: { active: true }
      })
  ).toHaveAuthorizationError()
})

test('boolean permission with null field is not authorized', async () => {
  const { object: post } = await actions
    .createWithBoolean({ active: false })

  expect(
    await actions
      .updateWithBooleanPermissionFromField({ 
        where: { id: post.id },
        values: { active: null }
      })
  ).toHaveAuthorizationError()
})

test('identity permission with correct identity in context is authorized', async () => {
  const { identityId } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user@keel.xyz',
    password: '1234'})

  const { object: identity } = await Identity.findOne({ id: identityId })

  const { object: post } = await actions
    .withIdentity(identity)
    .createWithIdentity({});

  expect(
    await actions
      .withIdentity(identity)
      .updateWithIdentityPermission({ 
        where: { id: post.id },
        values: { title: "hello" }
      })
  ).notToHaveAuthorizationError()
})

test('identity permission with incorrect identity in context is authorized', async () => {
  const { identityId: id1 } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user1@keel.xyz',
    password: '1234'})

  const { identityId: id2 } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user2@keel.xyz',
    password: '1234'})

  const { object: identity1 } = await Identity.findOne({ id: id1 })
  const { object: identity2 } = await Identity.findOne({ id: id2 })

  const { object: post } = await actions
    .withIdentity(identity1)
    .createWithIdentity({});

  expect(
    await actions
      .withIdentity(identity2)
      .updateWithIdentityPermission({ 
        where: { id: post.id },
        values: { title: "hello" }
      })
  ).toHaveAuthorizationError()
})

test('identity permission with no identity in context is not authorized', async () => {
  const { identityId: id } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user@keel.xyz',
    password: '1234'})

  const { object: identity } = await Identity.findOne({ id: id })

  const { object: post } = await actions
    .withIdentity(identity)
    .createWithIdentity({});

  expect(
    await actions
      .updateWithIdentityPermission({ 
        where: { id: post.id },
        values: { title: "hello" }
      })
  ).toHaveAuthorizationError()
})
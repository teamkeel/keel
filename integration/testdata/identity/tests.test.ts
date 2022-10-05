import { test, expect, actions, Post, Identity } from '@teamkeel/testing'

test('create identity', async () => {
  const { identityId, identityCreated } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user1@keel.xyz',
    password: '1234'})

  expect(identityCreated).toEqual(true)
})

test('do not create identity', async () => {
  const { identityId, identityCreated } = await actions.authenticate({ 
    createIfNotExists: false, 
    email: 'user2@keel.xyz',
    password: '1234'})

  expect(identityId).toBeEmpty()
  expect(identityCreated).toEqual(false)
})

test('authentication successful', async () => {
  const { identityId: id1, identityCreated: created1 } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user3@keel.xyz',
    password: '1234'})

  const { identityId: id2, identityCreated: created2 } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user3@keel.xyz',
    password: '1234'})

  expect(id1).toEqual(id2)
  expect(created1).toEqual(true)
  expect(created2).toEqual(false)
})

test('authentication unsuccessful', async () => {
  const { identityId: id1, identityCreated: created1, errors: errors1 } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user4@keel.xyz',
    password: '1234'})

  const { identityId: id2, identityCreated: created2 } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user4@keel.xyz',
    password: 'zzzz'})

  var notEqualIdentities = id1 != id2
  expect(notEqualIdentities).toEqual(true)
  expect(created1).toEqual(true)
  expect(created2).toEqual(false)
})

test('authorization successful', async () => {
  const { identityId } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user@keel.xyz',
    password: '1234'})

  const { object: identity } = await Identity.findOne({ id: identityId })

  const { object: post } = await actions
    .withIdentity(identity)
    .createPostWithIdentity({ title: 'temp' });

  expect(
    await actions
    .withIdentity(identity)
    .getPostRequiresIdentity({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('authorization failed', async () => {
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
    .createPostWithIdentity({ title: 'temp' });

  expect(
    await actions
    .withIdentity(identity2)
    .getPostRequiresIdentity({ id: post.id })
  ).toHaveAuthorizationError()
})

// todo:  permission test against null.  Requires this fix:  https://linear.app/keel/issue/DEV-195/permissions-support-null-operand-with-identity-type

// todo:  permission test against another identity field.  Requires this fix: https://linear.app/keel/issue/DEV-196/permissions-support-identity-type-operand-with-identity-comparison

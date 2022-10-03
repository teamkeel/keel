import { test, expect, actions, Post, Identity } from '@teamkeel/testing'

test('authorization successful', async () => {
  const { identityId } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user@keel.xyz',
    password: '1234'})

  const { object: identity } = await Identity.findOne({ id: identityId })

  const { object: post } = await actions
    .withIdentity(identity)
    .createPost({ title: 'temp' });

  expect(
    await actions
    .withIdentity(identity)
    .getPost({ id: post.id })
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
    .createPost({ title: 'temp' });

  expect(
    await actions
    .withIdentity(identity2)
    .getPost({ id: post.id })
  ).toHaveAuthorizationError()
})

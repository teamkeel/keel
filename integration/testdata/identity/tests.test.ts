import { test, expect, actions, Post } from '@teamkeel/testing'

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
import { test, expect, actions, Post, Identity, ModelWithExpressions } from '@teamkeel/testing'

test('permission set on model level for create op - matching title - is authorized', async () => {
  expect(
    await actions
      .create({ title: "hello" })
  ).notToHaveAuthorizationError()
})

test('permission set on model level for create op - not matching - is not authorized', async () => {
  expect(
    await actions
      .create({ title: "goodbye" })
  ).toHaveAuthorizationError()
})

test('ORed permissions set on model level for get op - matching title - is authorized', async () => {
  const { object: post } = await actions.create({ title: "hello", views: null })
  
  expect(
    await actions.get({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('ORed permissions set on model level for get op - matching title and views - is authorized', async () => {
  const { object: post } = await actions.create({ title: "hello", views: 5 })
  
  expect(
    await actions.get({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('ORed permissions set on model level for get op - none matching - is not authorized', async () => {
  const { object: post } = await actions.create({ title: "hello", views: 500 })
  
  await actions.update({ 
    where: { id: post.id },
    values: { title: "goodbye" }
  })

  expect(
    await actions.get({ id: post.id })
  ).toHaveAuthorizationError()
})

test('no permissions set on model level for delete op - can delete - is authorized', async () => {
  const { object: post } = await actions.create({ title: "hello", views: 500 })
  
  expect(
    await actions.delete({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('text literal comparisons - all expressions fail - is not authorized', async () => {
  expect(
    await actions.textsFailedExpressions({ title: "hello", explTitle: "hello" })
  ).toHaveAuthorizationError()
})

test('number literal comparisons - all expressions fail - is not authorized', async () => {
  expect(
    await actions.numbersFailedExpressions({ views: 2, explViews: 2 })
  ).toHaveAuthorizationError()
})

test('boolean literal comparisons - all expressions fail - is not authorized', async () => {
  expect(
    await actions.booleansFailedExpressions({ isActive: false, explIsActive: false })
  ).toHaveAuthorizationError()
})

test('enum literal comparisons - all expressions fail - is not authorized', async () => {
  expect(
    await actions.enumFailedExpressions({ option: "One", explOption: "One" })
  ).toHaveAuthorizationError()
})

test('permission role email is authorized', async () => {
  const { identityId } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'editorFred99@agency.org',
    password: '1234'
  })
  const { object: identity } = await Identity.findOne({ id: identityId })

  expect(
    await actions
      .withIdentity(identity)
      .createUsingRole({ title: "nothing special about this title" })
  ).notToHaveAuthorizationError()
})


test('permission role wrong email is not authorized', async () => {
  const { identityId } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'editorFred42@agency.org',
    password: '1234'
  })
  const { object: identity } = await Identity.findOne({ id: identityId })

  expect(
    await actions
      .withIdentity(identity)
      .createUsingRole({ title: "nothing special about this title" })
  ).toHaveAuthorizationError()
})


test('permission role domain is authorized', async () => {
  const { identityId } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'john.witherow@times.co.uk',
    password: '1234'
  })
  const { object: identity } = await Identity.findOne({ id: identityId })

  expect(
    await actions
      .withIdentity(identity)
      .createUsingRole({ title: "nothing special about this title" })
  ).notToHaveAuthorizationError()
})

test('permission role wrong domain is not authorized', async () => {
  const { identityId } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'jon.sargent@.bbc.co.uk',
    password: '1234'
  })
  const { object: identity } = await Identity.findOne({ id: identityId })

  expect(
    await actions
      .withIdentity(identity)
      .createUsingRole({ title: "nothing special about this title" })
  ).toHaveAuthorizationError()
})



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
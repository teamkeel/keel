import { test, expect, Actions, Thing } from '@teamkeel/testing'

test('do not set optional', async () => {
  const { object: createdThing } = await Actions.createThingWithoutOptionalNoDefault({})
  expect.equal(createdThing.optionalNoDefault, null)
})

test('set optional field with literal', async () => {
  const { object: createdThing } = await Actions.createThingWithExplicitSetOnOptionalNoDefaultField({})
  expect.equal(createdThing.optionalNoDefault, false)
})

test('set optional field with null', async () => {
  const { object: createdThing } = await Actions.createThingWithNullOnOptionalNoDefaultField({})
  expect.equal(createdThing.optionalNoDefault, null)
})

test('do not set optional field with default value', async () => {
  const { object: createdThing } = await Actions.createThingWithoutOnOptionalWithDefaultField({})
  expect.equal(createdThing.optionalWithDefault, true)
})

test('set optional field with default value with literal', async () => {
  const { object: createdThing } = await Actions.createThingWithExplicitOnOptionaWithDefaultField({})
  expect.equal(createdThing.optionalWithDefault, false)
})

test('set optional field with default value with null', async () => {
  const { object: createdThing } = await Actions.createThingWithNullOnOptionalWithDefaultField({})
  expect.equal(createdThing.optionalWithDefault, null)
})

test('do not set required field with default value', async () => {
  const { object: createdThing } = await Actions.createThingWithoutRequiredWithDefaultField({})
  expect.equal(createdThing.requiredWithDefault, true)
})

test('set required field with default value with literal', async () => {
  const { object: createdThing } = await Actions.createThingWithExplicitOnRequiredWithDefaultField({})
  expect.equal(createdThing.requiredWithDefault, false)
})
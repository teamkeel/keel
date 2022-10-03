import { test, expect, actions, Thing } from '@teamkeel/testing'

test('do not set optional', async () => {
  const { object: createdThing } = await actions
    .createThingWithoutOptionalNoDefault({})

  expect(createdThing.optionalNoDefault).toBeEmpty()
})

test('set optional field with literal', async () => {
  const { object: createdThing } = await actions
  .createThingWithExplicitSetOnOptionalNoDefaultField({})
  
  expect(createdThing.optionalNoDefault).toEqual(false)
})

test('set optional field with null', async () => {
  const { object: createdThing } = await actions
  .createThingWithNullOnOptionalNoDefaultField({})

  expect(createdThing.optionalNoDefault).toBeEmpty()
})

test('do not set optional field with default value', async () => {
  const { object: createdThing } = await actions
  .createThingWithoutOnOptionalWithDefaultField({})

  expect(createdThing.optionalWithDefault).toEqual(true)
})

test('set optional field with default value with literal', async () => {
  const { object: createdThing } = await actions
  .createThingWithExplicitOnOptionaWithDefaultField({})

  expect(createdThing.optionalWithDefault).toEqual(false)
})

test('set optional field with default value with null', async () => {
  const { object: createdThing } = await actions
  .createThingWithNullOnOptionalWithDefaultField({})

  expect(createdThing.optionalWithDefault).toBeEmpty()
})

test('do not set required field with default value', async () => {
  const { object: createdThing } = await actions
  .createThingWithoutRequiredWithDefaultField({})

  expect(createdThing.requiredWithDefault).toEqual(true)
})

test('set required field with default value with literal', async () => {
  const { object: createdThing } = await actions
  .createThingWithExplicitOnRequiredWithDefaultField({})

  expect(createdThing.requiredWithDefault).toEqual(false)
})
import { test, expect, actions, Thing } from '@teamkeel/testing'

test('do not set optional', async () => {
  const { object: createdThing } = await actions
    .createThingWithoutOptionalNoDefault({})

  expect.equal(createdThing.optionalNoDefault, null)
})

test('set optional field with literal', async () => {
  const { object: createdThing } = await actions
  .createThingWithExplicitSetOnOptionalNoDefaultField({})
  
  expect.equal(createdThing.optionalNoDefault, false)
})

test('set optional field with null', async () => {
  const { object: createdThing } = await actions
  .createThingWithNullOnOptionalNoDefaultField({})

  expect.equal(createdThing.optionalNoDefault, null)
})

test('do not set optional field with default value', async () => {
  const { object: createdThing } = await actions
  .createThingWithoutOnOptionalWithDefaultField({})

  expect.equal(createdThing.optionalWithDefault, true)
})

test('set optional field with default value with literal', async () => {
  const { object: createdThing } = await actions
  .createThingWithExplicitOnOptionaWithDefaultField({})

  expect.equal(createdThing.optionalWithDefault, false)
})

test('set optional field with default value with null', async () => {
  const { object: createdThing } = await actions
  .createThingWithNullOnOptionalWithDefaultField({})

  expect.equal(createdThing.optionalWithDefault, null)
})

test('do not set required field with default value', async () => {
  const { object: createdThing } = await actions
  .createThingWithoutRequiredWithDefaultField({})

  expect.equal(createdThing.requiredWithDefault, true)
})

test('set required field with default value with literal', async () => {
  const { object: createdThing } = await actions
  .createThingWithExplicitOnRequiredWithDefaultField({})

  expect.equal(createdThing.requiredWithDefault, false)
})
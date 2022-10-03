import { test, expect, actions, Thing } from '@teamkeel/testing'

test('do not set optional', async () => {
  const { object: createdThing } = await actions
  .createThingWithoutOptionalNoDefault({})
  
  expect(createdThing.optionalNoDefault).toBeEmpty()
})

test('set optional field with literal', async () => {
  const { object: createdThing } = await actions
  .createThingWithExplicitSetOnOptionalNoDefaultField({})

  expect(createdThing.optionalNoDefault).toEqual(99)
})

// TODO: https://linear.app/keel/issue/DEV-187/schema-parsing-error-with-expressions-of-negative-number-literal
// test('set optional field with negative literal', async () => {
//   const { object: createdThing } = await Actions.createThingWithNegativeOnOptionalNoDefaultField({})
//   expect(createdThing.optionalNoDefault).toEqual(-1)
// })

test('set optional field with null', async () => {
  const { object: createdThing } = await actions
  .createThingWithNullOnOptionalNoDefaultField({})

  expect(createdThing.optionalNoDefault).toBeEmpty()
})

test('do not set optional field with default value', async () => {
  const { object: createdThing } = await actions
  .createThingWithoutOnOptionalWithDefaultField({})

  expect(createdThing.optionalWithDefault).toEqual(1)
})

test('set optional field with default value with literal', async () => {
  const { object: createdThing } = await actions
  .createThingWithExplicitOnOptionaWithDefaultField({})

  expect(createdThing.optionalWithDefault).toEqual(99)
})

// TODO: https://linear.app/keel/issue/DEV-187/schema-parsing-error-with-expressions-of-negative-number-literal
// test('set optional field with default value with negative literal', async () => {
//   const { object: createdThing } = await Actions.createThingWithNegativeOnOptionalWithDefaultField({})
//   expect(createdThing.optionalWithDefault).toEqual(-1)
// })

test('set optional field with default value with null', async () => {
  const { object: createdThing } = await actions
  .createThingWithNullOnOptionalWithDefaultField({})

  expect(createdThing.optionalWithDefault).toBeEmpty()
})

test('do not set required field with default value', async () => {
  const { object: createdThing } = await actions
  .createThingWithoutRequiredWithDefaultField({})

  expect(createdThing.requiredWithDefault).toEqual(1)
})

test('set required field with default value with literal', async () => {
  const { object: createdThing } = await actions
  .createThingWithExplicitOnRequiredWithDefaultField({})

  expect(createdThing.requiredWithDefault).toEqual(99)
})

// TODO: https://linear.app/keel/issue/DEV-187/schema-parsing-error-with-expressions-of-negative-number-literal
// test('set required field with default value with negative literal', async () => {
//   const { object: createdThing } = await Actions.createThingWithNegativeOnRequiredWithDefaultField({})
//   expect(createdThing.requiredWithDefault).toEqual(-1)
// })
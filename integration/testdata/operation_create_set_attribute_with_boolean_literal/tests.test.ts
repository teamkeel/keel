import { test, expect, actions, Thing } from '@teamkeel/testing'

test('do not set optional', async () => {
  const { object: createdThing } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createThingWithoutOptionalNoDefault({})

  expect.equal(createdThing.optionalNoDefault, null)
})

test('set optional field with literal', async () => {
  const { object: createdThing } = await actions
  .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
  .createThingWithExplicitSetOnOptionalNoDefaultField({})
  
  expect.equal(createdThing.optionalNoDefault, false)
})

test('set optional field with null', async () => {
  const { object: createdThing } = await actions
  .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
  .createThingWithNullOnOptionalNoDefaultField({})

  expect.equal(createdThing.optionalNoDefault, null)
})

test('do not set optional field with default value', async () => {
  const { object: createdThing } = await actions
  .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
  .createThingWithoutOnOptionalWithDefaultField({})

  expect.equal(createdThing.optionalWithDefault, true)
})

test('set optional field with default value with literal', async () => {
  const { object: createdThing } = await actions
  .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
  .createThingWithExplicitOnOptionaWithDefaultField({})

  expect.equal(createdThing.optionalWithDefault, false)
})

test('set optional field with default value with null', async () => {
  const { object: createdThing } = await actions
  .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
  .createThingWithNullOnOptionalWithDefaultField({})

  expect.equal(createdThing.optionalWithDefault, null)
})

test('do not set required field with default value', async () => {
  const { object: createdThing } = await actions
  .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
  .createThingWithoutRequiredWithDefaultField({})

  expect.equal(createdThing.requiredWithDefault, true)
})

test('set required field with default value with literal', async () => {
  const { object: createdThing } = await actions
  .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
  .createThingWithExplicitOnRequiredWithDefaultField({})

  expect.equal(createdThing.requiredWithDefault, false)
})
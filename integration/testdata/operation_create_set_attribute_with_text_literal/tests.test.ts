import { test, expect, actions, Thing } from '@teamkeel/testing'

test('do not set optional', async () => {
  const { object: createdThing } = await actions
  .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
  .createThingWithoutNickname({})

  expect.equal(createdThing.optionalNoDefault, null)
})

test('set optional field with literal', async () => {
  const { object: createdThing } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createThingWithExplicitSetOnOptionalNoDefaultField({})

  expect.equal(createdThing.optionalNoDefault, "explicit")
})

test('set optional field with empty literal', async () => {
  const { object: createdThing } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createThingWithEmptyOnOptionalNoDefaultField({})

  expect.equal(createdThing.optionalNoDefault, "")
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

  expect.equal(createdThing.optionalWithDefault, "default")
})

test('set optional field with default value with literal', async () => {
  const { object: createdThing } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createThingWithExplicitOnOptionaWithDefaultField({})

  expect.equal(createdThing.optionalWithDefault, "explicit")
})

test('set optional field with default value with empty literal', async () => {
  const { object: createdThing } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createThingWithEmptyOnOptionalWithDefaultField({})

  expect.equal(createdThing.optionalWithDefault, "")
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

  expect.equal(createdThing.requiredWithDefault, "default")
})

test('set required field with default value with literal', async () => {
  const { object: createdThing } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createThingWithExplicitOnRequiredWithDefaultField({})

  expect.equal(createdThing.requiredWithDefault, "explicit")
})

test('set required field with default value with empty literal', async () => {
  const { object: createdThing } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createThingWithEmptyOnRequiredWithDefaultField({})
    
  expect.equal(createdThing.requiredWithDefault, "")
})
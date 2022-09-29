import { test, expect, actions, Thing } from '@teamkeel/testing'

test('allows for two list operations on same model', async () => {
  await Thing.create({ something: '123' })

  const { collection: one } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .listOne({})

  const { collection: two } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .listTwo({})

  expect.equal(one, two)
})

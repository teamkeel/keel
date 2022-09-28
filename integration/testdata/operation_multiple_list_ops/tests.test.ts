import { test, expect, Actions, Thing } from '@teamkeel/testing'

test('allows for two list operations on same model', async () => {
  await Thing.create({ something: '123' })
  const { collection: one } = await Actions.listOne({})
  const { collection: two } = await Actions.listTwo({})
  expect.equal(one, two)
})

import { test, expect, Actions } from '@teamkeel/testing'

test('it passes', async () => {
  const result = await Actions.createPerson({ title: 'something' })
  console.log(result)
  expect.equal(1, 1)
})

test('it fails', async () => {
  expect.equal(1, 2)
})

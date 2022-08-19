import { test, expect } from '@teamkeel/testing'

test('it passes', async () => {
  // const p = await Person.create({ title: 'test '})
  expect.equal(1, 1)
})

test('it fails', () => {
  expect.equal(1, 2)
})

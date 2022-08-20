import { test, expect, Actions } from '@teamkeel/testing'

test('it passes', async () => {
  console.log(Actions)
  expect.equal(1, 1)
})

test('it fails', () => {
  expect.equal(1, 2)
})

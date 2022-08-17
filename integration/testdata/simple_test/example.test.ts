import { test, expect } from '@teamkeel/testing'

test('it passes', () => {
  expect.equal(1, 1)
})

test('it fails', () => {
  expect.equal(1, 2)
})

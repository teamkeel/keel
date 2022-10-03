import { test, expect } from '@teamkeel/testing'

test('it passes', () => {
  expect(1).toEqual(1)
})

test('it fails', () => {
  expect(1).toEqual(2)
})

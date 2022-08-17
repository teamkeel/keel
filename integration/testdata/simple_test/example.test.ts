import { test, expect, runAllTests } from '@teamkeel/testing'

test('it passes', () => {
  expect.equal(1, 1)
})

test('it fails', () => {
  expect.equal(1, 2)
})

runAllTests({ parentPort: parseInt( process.env.HOST_PORT, 10) })
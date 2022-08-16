import { test, expect, runAllTests } from '@teamkeel/testing'

test('something', () => {
  expect.equal(1,1)
})

runAllTests({ parentPort: parseInt( process.env.HOST_PORT, 10) })
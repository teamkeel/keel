import { test, expect, runAllTests } from '../src/index'

test('a failing test', () => {
  expect.equal(2, 1)
  expect.equal(2, 4)
})

test('a passing test', () => {
  expect.equal(2, 2)
})

runAllTests({ parentPort: 3000 })
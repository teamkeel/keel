import { RunnerOpts, Test, TestFunc, TestName } from './types'
import { AssertionFailure } from './errors'
import { TestCaseResult, TestResult } from './output'
import { expect } from './expect'
import Reporter from './reporter'

const tests : Test[] = []

function test(testName: TestName, fn: TestFunc) {
  tests.push({
    testName,
    fn,
  })
}

// global - reset with every instantiation of module.
let results: TestCaseResult[] = []

function runAllTests({ parentPort }: RunnerOpts) {
  const reporter = new Reporter({
    host: 'localhost',
    port: parentPort
  })
  results = []

  if (!tests.length) {
    return
  } 

  for (const { testName, fn } of tests) {
    let result = null

    try {
      fn()

      result = TestResult.pass(testName)
    } catch (err) {
      if (err instanceof AssertionFailure) {
        const { actual, expected } = err as AssertionFailure

        result = TestResult.fail(
          testName,
          actual,
          expected,
        )
      } else {
        result = TestResult.exception(testName, err as Error)
      }
    } finally {
      results.push(result!.asObject())
    }
  }

  // report back to parent process with all
  // results for tests in the current file.
  reporter.report(results)
}

export {
  test,
  expect,
  runAllTests
}
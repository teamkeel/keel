import { RunnerOpts, Test, TestFunc, TestName } from './types'
import { AssertionFailure } from './errors'
import { TestResultData, TestResult } from './output'
import { expect } from './expect'
import Reporter from './reporter'

// generated.ts doesnt exist at this point, but once the node_module has been
// injected with the generated code, IT WILL 😈
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
//@ts-ignore
export * from './generated'

const tests : Test[] = []

function test(testName: TestName, fn: TestFunc) {
  tests.push({
    testName,
    fn,
  })
}

// global - reset with every instantiation of module.
let results: TestResultData[] = []

async function runAllTests({ parentPort, host = 'localhost' }: RunnerOpts) {
  const reporter = new Reporter({
    host,
    port: parentPort
  })
  results = []

  if (!tests.length) {
    return
  } 

  for (const { testName, fn } of tests) {
    let result : TestResult | undefined = undefined

    try {
      const t = fn()

      // support both async and non async invocations:
      // i.e
      // test('a non async test', () => {})
      // and
      // test('an async test', async () => {})
      // will both be supported.
      const isPromisified = t instanceof Promise

      // if we do not await the result of the func,
      // then the catch block will not catch the error
      if (isPromisified) {
        await t
      }

      result = TestResult.pass(testName)
    } catch (err) {
      // If the above code throws, then we know something went wrong during execution
      // An AssertionFailure might have been thrown, but it could also be something
      // else, so we need to check with instance_of checks the type of error

      const isAssertionFailure = err instanceof AssertionFailure

      if (isAssertionFailure) {
        const { actual, expected } = err as AssertionFailure

        result = TestResult.fail(
          testName,
          actual,
          expected,
        )
      } else if (err instanceof Error) {
        result = TestResult.exception(testName, err)
      } else {
        // if it's not an error, then wrap after stringifing
        result = TestResult.exception(testName, new Error(JSON.stringify(err)))
      }
    } finally {
      if (result) {
        results.push(result.asObject())
      }
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

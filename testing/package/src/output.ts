import { TestName } from 'types'

enum Status {
  Pass = 'pass',
  Fail = 'fail',
  Skipped = 'skipped',
  Exception = 'exception'
}

export interface TestCaseResult {
  testName: TestName;
  status: Status;
  expected?: any;
  actual?: any;
}

export class TestResult {
  private readonly testName: TestName
  private readonly status: Status
  private readonly actual: any
  private readonly expected: any

  private constructor(status: Status, testName: string, actual?: any, expected?: any, err?: Error) {
    this.testName = testName
    this.status = status
    
    if (expected && actual) {
      this.actual = actual
      this.expected = expected
    }
  }

  static fail(testCase: string, actual: any, expected: any) {
    return new TestResult(Status.Fail, testCase, actual, expected)
  }

  static exception(testCase: string, err: Error) {
    return new TestResult(Status.Exception, testCase, undefined, undefined, err)
  }

  static pass(testCase: string) {
    return new TestResult(Status.Pass, testCase)
  }

  asObject = () : TestCaseResult => {
    let base: TestCaseResult = {
      testName: this.testName,
      status: this.status
    }

    if (this.expected && this.actual) {
      base = { ...base, expected: this.expected, actual: this.actual }
    }

    return base
  }

  toJSON = () => JSON.stringify(this.asObject())
}

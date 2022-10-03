import { TestName } from "./types";

enum Status {
  Pass = "pass",
  Fail = "fail",
  Skipped = "skipped",
  Exception = "exception",
}

export interface TestResultData {
  status: Status;
  testName: string;
  actual?: unknown;
  expected?: unknown;
  err?: Error;
}

export class TestResult {
  private readonly testName: TestName;
  private readonly status: Status;
  private readonly actual?: unknown;
  private readonly expected?: unknown;
  private readonly err?: Error;

  private constructor({
    testName,
    status,
    err,
    expected,
    actual,
  }: TestResultData) {
    this.testName = testName;
    this.status = status;
    if (err) {
      this.err = err;
    }

    if (expected && actual) {
      this.actual = actual;
      this.expected = expected;
    }
  }

  static fail(
    testName: string,
    actual: unknown,
    expected: unknown
  ): TestResult {
    return new TestResult({ status: Status.Fail, testName, actual, expected });
  }

  static exception(testName: string, err: Error): TestResult {
    return new TestResult({ status: Status.Exception, testName, err });
  }

  static pass(testName: string): TestResult {
    return new TestResult({ status: Status.Pass, testName });
  }

  asObject = (): TestResultData => {
    let base: TestResultData = {
      testName: this.testName,
      status: this.status,
    };

    if (this.expected && this.actual) {
      base = { ...base, expected: this.expected, actual: this.actual };
    }

    if (this.err) {
      base = { ...base, err: this.err };
    }

    return base;
  };

  toJSON = () => {
    const obj = this.asObject();

    if (obj.err) {
      // Error instances are not automatically stringified to JSON
      // so we need to serialize their properties
      // See https://stackoverflow.com/questions/18391212/is-it-not-possible-to-stringify-an-error-using-json-stringify
      const { stack, message, name } = obj.err;
      return JSON.stringify({
        ...obj,
        err: {
          message,
          stack,
          name,
        },
      });
    }

    return JSON.stringify(this.asObject());
  };
}

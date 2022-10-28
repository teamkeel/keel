import { TestCase } from "./types";

enum Status {
  Pass = "pass",
  Fail = "fail",
  Skipped = "skipped",
  Exception = "exception",
}

export interface TestResultData {
  status: Status;
  test: TestCase;
  actual?: unknown;
  expected?: unknown;
  err?: Error;
}

export class TestResult {
  private readonly test: TestCase;
  private readonly status: Status;
  private readonly actual?: unknown;
  private readonly expected?: unknown;
  private readonly err?: Error;

  private constructor({
    test,
    status,
    err,
    expected,
    actual,
  }: TestResultData) {
    this.test = test;
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
    test: TestCase,
    actual: unknown,
    expected: unknown
  ): TestResult {
    return new TestResult({ status: Status.Fail, test, actual, expected });
  }

  static exception(test: TestCase, err: Error): TestResult {
    return new TestResult({ status: Status.Exception, test, err });
  }

  static pass(test: TestCase): TestResult {
    return new TestResult({ status: Status.Pass, test });
  }

  asObject = (): TestResultData => {
    let base: TestResultData = {
      test: this.test,
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
      return {
        ...obj,
        err: {
          message,
          stack,
          name,
        },
      };
    }

    return this.asObject();
  };
}

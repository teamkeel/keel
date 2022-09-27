import { Logger, LogLevel } from "@teamkeel/sdk";
import chalk from "chalk";
import { DatabaseError } from "pg-protocol";

import { RunnerOpts, Test, TestFunc, TestName } from "./types";
import { AssertionFailure } from "./errors";
import { TestResultData, TestResult } from "./output";
import { expect } from "./expect";
import Reporter from "./reporter";

// generated.ts doesnt exist at this point, but once the node_module has been
// injected with the generated code, IT WILL 😈
//@ts-ignore
export * from "./generated";

const runnerLogger = new Logger({ colorize: true });

const tests: Test[] = [];

function test(testName: TestName, fn: TestFunc) {
  tests.push({
    testName,
    fn,
  });
}

// global - reset with every instantiation of module.
let results: TestResultData[] = [];

async function runAllTests({
  parentPort,
  host = "localhost",
  debug,
  filePath,
  pattern = "",
}: RunnerOpts) {
  console.log(`${chalk.white.bgBlue(" INFO ")} Running testfile ${filePath}\n`);

  const hasPattern = pattern !== "";

  if (hasPattern) {
    console.log(`${chalk.white.bgBlue(" INFO ")} Filtering on ${pattern}\n`);
  }

  const reporter = new Reporter({
    host,
    port: parentPort,
  });
  results = [];

  if (!tests.length) {
    return;
  }

  for (const { testName, fn } of tests) {
    if (hasPattern) {
      const regex = new RegExp(pattern!);

      if (!regex.test(testName)) {
        continue;
      }

      console.log(`${chalk.bgYellow.white(" RUNS ")} ${testName}\n`);
    } else {
      console.log(`${chalk.bgYellow.white(" RUNS ")} ${testName}\n`);
    }

    let result: TestResult | undefined = undefined;

    try {
      const t = fn();

      // support both async and non async invocations:
      // i.e
      // test('a non async test', () => {})
      // and
      // test('an async test', async () => {})
      // will both be supported.
      const isPromisified = t instanceof Promise;

      // if we do not await the result of the func,
      // then the catch block will not catch the error
      if (isPromisified) {
        await t;
      }

      result = TestResult.pass(testName);

      console.log(`${chalk.bgGreen.white(" PASS ")} ${testName}\n`);
    } catch (err) {
      if (debug) {
        console.debug(err);
      }

      // If the above code throws, then we know something went wrong during execution
      // An AssertionFailure might have been thrown, but it could also be something
      // else, so we need to check with instance_of checks the type of error

      const isAssertionFailure = err instanceof AssertionFailure;

      if (isAssertionFailure) {
        const { actual, expected } = err as AssertionFailure;

        result = TestResult.fail(testName, actual, expected);

        console.log(`${chalk.bgRed.white(" FAIL ")} ${testName}\n`);
      } else if (err instanceof DatabaseError) {
        // do nothing

        console.log(
          `${chalk.bgBlueBright.white(
            " INFO "
          )} Connection terminated during execution of ${testName}\n`
        );
      } else if (err instanceof Error) {
        // An unrelated error occurred inside of the .test() block
        // which was an instanceof Error
        result = TestResult.exception(testName, err);
        console.log(`${chalk.bgRedBright.white(" ERROR ")} ${testName}\n`);
        runnerLogger.log(`${err}\n${err.stack}`, LogLevel.Error);
      } else {
        // if it's not an error, then wrap after stringifing
        result = TestResult.exception(testName, new Error(JSON.stringify(err)));
        console.log(`${chalk.bgRedBright.white(" ERROR ")} ${testName}\n`);
        runnerLogger.log(`${err}`, LogLevel.Error);
      }
    } finally {
      if (result) {
        if (debug) {
          console.debug(result.asObject());
        }

        results.push(result.asObject());
      }
    }

    const resetSuccess = await reporter.clearDatabase();

    if (debug) {
      if (resetSuccess) {
        console.log(
          `${chalk.bgBlueBright.white(
            " INFO "
          )} Reset database after ${testName}\n`
        );
      } else {
        console.log(
          `${chalk.bgRedBright.white(
            " ERROR "
          )} Could not reset database after ${testName}\n`
        );
      }
    }
  }

  // report back to parent process with all
  // results for tests in the current file.
  // we want to await for it to complete prior to moving on
  // because the report request also clears the database
  // between individual test() cases
  await reporter.report(results);
}

const logger = new Logger({ colorize: true, timestamps: false });

process.on("uncaughtException", () => {
  return;
});

export { test, expect, runAllTests, logger, LogLevel };

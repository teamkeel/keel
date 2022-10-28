import { Logger, LogLevel } from "@teamkeel/sdk";
import chalk from "chalk";
import { DatabaseError } from "pg-protocol";

import { RunnerOpts, Test, TestFunc, TestName } from "./types";
import { AssertionFailure } from "./errors";
import { TestResult } from "./output";
import { expect } from "./expect";
import Reporter from "./reporter";
import log from './logger'

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

async function runAllTests({
  parentPort,
  host = "localhost",
  debug,
  filePath,
  pattern = "",
  silent = false
}: RunnerOpts) {
  log(`${chalk.white.bgBlue(" INFO ")} Running testfile ${filePath}\n`, silent);

  const hasPattern = pattern !== "";

  if (hasPattern) {
    log(`${chalk.white.bgBlue(" INFO ")} Filtering on ${pattern}\n`, silent);
  }

  const reporter = new Reporter({
    host,
    port: parentPort,
  });

  if (!tests.length) {
    return;
  }

  for (const { testName, fn } of tests) {
    if (hasPattern) {
      const regex = new RegExp(pattern!);

      if (!regex.test(testName)) {
        continue;
      }

      log(`${chalk.bgYellow.white(" RUNS ")} ${testName}\n`, silent);
    } else {
      log(`${chalk.bgYellow.white(" RUNS ")} ${testName}\n`, silent);
    }

    let result: TestResult | undefined = undefined;

    // we make a http request to the /reset endpoint in the go process
    // which resets the database prior to the test run
    const resetSuccess = await reporter.clearDatabase({
      filePath: filePath,
      testCase: testName,
    });

    if (debug) {
      if (resetSuccess) {
        log(
          `${chalk.bgBlueBright.white(
            " INFO "
          )} Reset database after ${testName}\n`,
          silent
        );
      } else {
        log(
          `${chalk.bgRedBright.white(
            " ERROR "
          )} Could not reset database after ${testName}\n`,
          silent
        );
      }
    }

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

      log(`${chalk.bgGreen.white(" PASS ")} ${testName}\n`, silent);
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

        log(`${chalk.bgRed.white(" FAIL ")} ${testName}\n`, silent);
      } else if (err instanceof DatabaseError) {
        // do nothing

        log(
          `${chalk.bgBlueBright.white(
            " INFO "
          )} Connection terminated during execution of ${testName}\n`,
          silent
        );
      } else if (err instanceof Error) {
        // An unrelated error occurred inside of the .test() block
        // which was an instanceof Error
        result = TestResult.exception(testName, err);
        log(`${chalk.bgRedBright.white(" ERROR ")} ${testName}\n`, silent);
        runnerLogger.log(`${err}\n${err.stack}`, LogLevel.Error);
      } else {
        // if it's not an error, then wrap after stringifing
        result = TestResult.exception(testName, new Error(JSON.stringify(err)));
        log(`${chalk.bgRedBright.white(" ERROR ")} ${testName}\n`, silent);
        runnerLogger.log(`${err}`, LogLevel.Error);
      }
    } finally {
      if (result) {
        if (debug) {
          console.debug(result.asObject());
        }

        // report back to parent process with
        // result for test
        await reporter.report([result]);
      }
    }
  }
}

const logger = new Logger({ colorize: true, timestamps: false });

// todo: replace this logic with more graceful termination
// logic
// A better way might be to refactor the testing.go file
// so that it creates a database per *JS test case* although
// this requires the go world to have knowledge of each
// of the individual test cases.

// Explanation:
// We call pg_terminate_backend in the go test harness
// to terminate any active connections to the database
// when clearing the database between individual test cases
// This causes any active connections to the db in the node process
// to error catastrophically.
// which doesnt seem to be caught by standard try/catch
// mechanism. Only process.on('uncaughtException') seems
// to catch the error so we just want to return early for this
// case instead of exiting

// Postgres docs: https://www.postgresql.org/docs/8.0/errcodes-appendix.html#:~:text=57P01,ADMIN%20SHUTDOWN
const ADMIN_SHUTDOWN_CODE = "57P01";

process.on("uncaughtException", (err, next) => {
  const { name } = err.constructor;

  // err instanceof DatabaseError doesnt work when the DatabaseError is from a different package
  if (
    name === "DatabaseError" &&
    (err as DatabaseError).code === ADMIN_SHUTDOWN_CODE
  ) {
    return;
  }

  // If it's any other kind of uncaught exception then exit with a non zero exit code
  console.log("Exiting process due to", err);
  process.exit(1);
});

export { test, expect, runAllTests, logger, LogLevel };

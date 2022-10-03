import isEqual from "lodash.isequal";

import { ActionError, ActionResult, Matchers, Value } from "./types";
import { AssertionFailure } from "./errors";

type Expect = (actual: Value) => Matchers;

const expect: Expect = (actual: Value) => ({
  toEqual: (expected: any): void => {
    // Lodash's isEqual checks equality for many different types:
    // arrays, booleans, dates, errors, maps, numbers, objects, regexes, sets, strings, symbols
    // Ref: https://lodash.com/docs/4.17.15#isEqual
    if (!isEqual(actual, expected)) {
      // We throw an AssertionFailure error here
      // This error is caught by a try/catch in the test method
      // And the test failure is reported back to the golang reporting server.
      throw new AssertionFailure(actual, expected);
    }
  },
  notToEqual: (expected: Value): void => {
    if (isEqual(actual, expected)) {
      // We throw an AssertionFailure error here
      // This error is caught by a try/catch in the test method
      // And the test failure is reported back to the golang reporting server.
      throw new AssertionFailure(actual, expected);
    }
  },
  // todo: narrow error type below after sdk publish
  toHaveError: (error: ActionError): void => {
    const actionResult = actual as unknown as ActionResult;

    if (!actionResult.errors) {
      throw new AssertionFailure(null, error);
    }

    const match = actionResult.errors.find((e) => {
      // todo: this needs to be formalised once
      // error payload is developed further.
      // e.g factor in error code etc
      return e.message === error.message;
    });

    if (!match) {
      throw new AssertionFailure(
        actionResult.errors,
        actionResult.errors.concat(error)
      );
    }
  },
  // Checks for both null and undefined
  toBeEmpty: (): void => {
    if (actual !== undefined && actual !== null) {
      throw new AssertionFailure(actual, null);
    }
  },
  // Checks the actual value isnt null or undefined
  notToBeEmpty: (): void => {
    if (actual === undefined || actual === null) {
      throw new AssertionFailure(actual, actual);
    }
  },
  // toContain will check for existence of item in actual array
  toContain: (item: Value): void => {
    if (!Array.isArray(actual)) {
      throw new Error("actual is not an array");
    }

    const match = actual.find((i) => isEqual(i, item));

    if (!match) {
      throw new AssertionFailure(actual, actual.concat(item));
    }
  },
  notToContain: (item: Value): void => {
    if (!Array.isArray(actual)) {
      throw new Error("actual is not an array");
    }
    const match = actual.find((i) => isEqual(i, item));

    if (match) {
      throw new AssertionFailure(actual, actual.concat(item));
    }
  },
});

export { expect };

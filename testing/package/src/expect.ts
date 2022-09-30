import isEqual from "lodash.isequal";

import { Matchers, Value, AssertionFunction } from './types'
import { AssertionFailure } from "./errors";

type Expect = (actual: Value) => Matchers

const expect : Expect = (actual: Value) => ({
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
  notToEqual: (expected: Value) : void => {
    if (isEqual(actual, expected)) {
      // We throw an AssertionFailure error here
      // This error is caught by a try/catch in the test method
      // And the test failure is reported back to the golang reporting server.
      throw new AssertionFailure(actual, expected);
    }
  },
  // todo: narrow error type below after sdk publish
  toHaveError: (error: Value) : void => {

  },
  // Checks for both null and undefined
  toBeEmpty: () : void => {

  },
  // Checks the actual value isnt null or undefined
  notToBeEmpty: () : void => {

  },
  toBeTrue: () : void => {

  },
  toBeFalse: () : void => {

  },
  // toMatchArray will check that actual and expected contain the same elements
  // without worrying about order.
  toMatchArray: () : void => {

  },
  // toContain will check for existence of item in actual array
  toContain: () : void => {

  },
  satisfy: (assertionFunc: AssertionFunction) : void => {
    if (!assertionFunc(actual)) {
      throw new AssertionFailure(actual)
    }
  }
})

export { expect };

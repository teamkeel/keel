import isEqual from "lodash.isequal";
import { AssertionFailure } from "./errors";

const expect = {
  // todo: narrow expected + actual type from any to something serializable
  equal: (actual: any, expected: any): void => {
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
};

export { expect };

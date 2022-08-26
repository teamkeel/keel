import isEqual from 'lodash.isequal'
import { AssertionFailure } from "./errors";

const expect = {
  equal: (actual: any, expected: any): void => {
    // Lodash's isEqual checks equality for many different types:
    // arrays, booleans, dates, errors, maps, numbers, objects, regexes, sets, strings, symbols
    // Ref: https://lodash.com/docs/4.17.15#isEqual
    if (!isEqual(actual, expected)) {
      throw new AssertionFailure(actual, expected);
    }
  },
};

export { expect };

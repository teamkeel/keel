import { AssertionFailure } from './errors'

const expect = {
  equal: (actual: any, expected: any) : void => {
    if (actual !== expected) {
      throw new AssertionFailure(actual, expected)
    }
  }
}

export { expect }

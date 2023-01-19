import isMatch from 'lodash.ismatch'

export async function toHaveError(received, expected) {
  const { isNot } = this;
  try {
    const v = await received;

    return {
      pass: false,
      message: () => 'expected value to reject',
      actual: JSON.stringify(v),
      expected: JSON.stringify(expected)
    };
  } catch (err) {
    return {
      pass: isMatch(err, expected),
      message: () =>
        `expected ${isNot ? "no " : ""} ${JSON.stringify(err)} error`,
      actual: JSON.stringify(err),
      expected: JSON.stringify(expected)
    };
  }
}

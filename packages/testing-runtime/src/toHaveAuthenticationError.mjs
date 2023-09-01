export async function toHaveAuthenticationError(received) {
  const { isNot } = this;
  try {
    const v = await received;
    return {
      pass: false,
      message: () => "expected value to reject",
      actual: v,
    };
  } catch (err) {
    return {
      pass: err.code === "ERR_AUTHENTICATION_FAILED",
      message: () =>
        `expected there to be ${
          isNot ? "no " : ""
        }ERR_AUTHENTICATION_FAILED error`,
      actual: err,
      expected: {
        ...err,
      },
    };
  }
}

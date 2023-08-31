export async function toHaveAuthorizationError(received) {
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
      pass: err.code === "ERR_PERMISSION_DENIED",
      message: () =>
        `expected there to be ${isNot ? "no " : ""}ERR_PERMISSION_DENIED error`,
      actual: err,
      expected: {
        ...err,
      },
    };
  }
}

export async function toHaveAuthorizationError(received) {
  const { isNot } = this;
  try {
    const v = await received;
    return {
      pass: false,
      message: () => "expected value to reject",
      actual: v,
      expected: {
        code: "ERR_PERMISION_DENIED",
      },
    };
  } catch (err) {
    return {
      pass: err.code === "ERR_PERMISSION_DENIED",
      message: () =>
        `expected ${isNot ? "no " : ""}ERR_PERMISSION_DENIED error`,
      actual: err,
      expected: {
        code: "ERR_PERMISION_DENIED",
      },
    };
  }
}

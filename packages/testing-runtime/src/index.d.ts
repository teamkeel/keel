// See https://vitest.dev/guide/extending-matchers.html for docs
// on typing custom matchers

interface ActionError {
  code: string;
  message: string;
}

interface CustomMatchers<R = unknown> {
  toHaveAuthorizationError(): void;
  toHaveAuthenticationError(): void;
  toHaveError(err: Partial<ActionError>): void;
}

declare global {
  namespace Vi {
    interface Assertion extends CustomMatchers {}
    interface AsymmetricMatchersContaining extends CustomMatchers {}
  }
}

export {};

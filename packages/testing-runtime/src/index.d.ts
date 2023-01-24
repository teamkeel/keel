// See https://vitest.dev/guide/extending-matchers.html for docs
// on typing custom matchers

interface ActionError {
  code: string;
}

interface CustomMatchers<R = unknown> {
  toHaveAuthorizationError(): void;
  toHaveError(err: ActionError): void;
}

declare global {
  namespace Vi {
    interface Assertion extends CustomMatchers {}
    interface AsymmetricMatchersContaining extends CustomMatchers {}
  }
}

export {};

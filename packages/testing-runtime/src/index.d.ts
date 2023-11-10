// See https://vitest.dev/guide/extending-matchers.html for docs
// on typing custom matchers
import type { Assertion, AsymmetricMatchersContaining } from "vitest";

interface ActionError {
  code: string;
  message: string;
}

interface CustomMatchers<R = unknown> {
  toHaveAuthorizationError(): void;
  toHaveAuthenticationError(): void;
  toHaveError(err: Partial<ActionError>): void;
}

declare module "vitest" {
  interface Assertion<T = any> extends CustomMatchers<T> {}
  interface AsymmetricMatchersContaining extends CustomMatchers {}
}

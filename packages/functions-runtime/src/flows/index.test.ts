import { test, expect } from "vitest";
import { RetryBackoffExponential, RetryBackoffLinear, RetryConstant } from ".";

test("retry delay - exponential backoffs", () => {
  // exponential backoff with base 2s - attempt 2 - first retry
  expect(RetryBackoffExponential(2)(2)).toBe(2000);
  // exponential backoff with base 2s - attempt 3 - second retry
  expect(RetryBackoffExponential(2)(3)).toBe(4000);
  // exponential backoff with base 2s - attempt 4 - third retry
  expect(RetryBackoffExponential(2)(4)).toBe(8000);
  // exponential backoff with base 2s - attempt 5 - fourth retry
  expect(RetryBackoffExponential(2)(5)).toBe(16000);
});

test("retry delay - linear backoffs", () => {
  // linear backoff with interval 3s - attempt 2
  expect(RetryBackoffLinear(3)(2)).toBe(3000);
  // linear backoff with interval 3s - attempt 3
  expect(RetryBackoffLinear(3)(3)).toBe(6000);
  // linear backoff with interval 3s - attempt 4
  expect(RetryBackoffLinear(3)(4)).toBe(9000);
});

test("retry delay - constant", () => {
  // constant delay 5s - attempt 2
  expect(RetryConstant(5)(2)).toBe(5000);
  // constant delay 5s - attempt 3
  expect(RetryConstant(5)(3)).toBe(5000);
  // constant delay 5s - attempt 4
  expect(RetryConstant(5)(4)).toBe(5000);
});

import { test, expect } from "vitest";
import { RetryBackoffExponential, RetryBackoffLinear, RetryConstant } from ".";

test("retry delay - exponential backoffs", () => {
  // exponential backoff with base 2s - retry 0
  expect(RetryBackoffExponential(2)(0)).toBe(0);
  // exponential backoff with base 2s - retry 1
  expect(RetryBackoffExponential(2)(1)).toBe(2000);
  // exponential backoff with base 2s - retry 2
  expect(RetryBackoffExponential(2)(2)).toBe(4000);
  // exponential backoff with base 2s - retry 3
  expect(RetryBackoffExponential(2)(3)).toBe(8000);
  // exponential backoff with base 2s - retry 4
  expect(RetryBackoffExponential(2)(4)).toBe(16000);
});

test("retry delay - linear backoffs", () => {
  // linear backoff with interval 3s - retry 0
  expect(RetryBackoffLinear(3)(0)).toBe(0);
  // linear backoff with interval 3s - retry 1
  expect(RetryBackoffLinear(3)(1)).toBe(3000);
  // linear backoff with interval 3s - retry 2
  expect(RetryBackoffLinear(3)(2)).toBe(6000);
  // linear backoff with interval 3s - retry 3
  expect(RetryBackoffLinear(3)(3)).toBe(9000);
  // linear backoff with interval 3s - retry 4
  expect(RetryBackoffLinear(3)(4)).toBe(12000);
});

test("retry delay - constant", () => {
  // constant delay 5s - retry 0
  expect(RetryConstant(5)(0)).toBe(0);
  // constant delay 5s - retry 1
  expect(RetryConstant(5)(1)).toBe(5000);
  // constant delay 5s - retry 2
  expect(RetryConstant(5)(2)).toBe(5000);
  //   constant delay 5s - retry 3
  expect(RetryConstant(5)(3)).toBe(5000);
  // constant delay 5s - retry 4
  expect(RetryConstant(5)(4)).toBe(5000);
});

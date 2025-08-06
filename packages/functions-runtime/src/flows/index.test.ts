import { test, expect } from "vitest";
import { RetryBackoffExponential, RetryBackoffLinear, RetryConstant } from ".";

test("retry delay - exponential backoffs", () => {
  // base 2s - retry 0
  expect(RetryBackoffExponential(2)(0)).toBe(0);
  // base 2s - retry 1
  expect(RetryBackoffExponential(2)(1)).toBe(2000);
  // base 2s - retry 2
  expect(RetryBackoffExponential(2)(2)).toBe(4000);
  // base 2s - retry 3
  expect(RetryBackoffExponential(2)(3)).toBe(8000);
  // base 2s - retry 4
  expect(RetryBackoffExponential(2)(4)).toBe(16000);

  // base 5s - retry 0
  expect(RetryBackoffExponential(5)(0)).toBe(0);
  // base 5s - retry 1
  expect(RetryBackoffExponential(5)(1)).toBe(5000);
  // base 5s - retry 2
  expect(RetryBackoffExponential(5)(2)).toBe(25000);
  // base 5s - retry 3
  expect(RetryBackoffExponential(5)(3)).toBe(125000);
  // base 5s - retry 4
  expect(RetryBackoffExponential(5)(4)).toBe(625000);
});

test("retry delay - linear backoffs", () => {
  // interval 3s - retry 0
  expect(RetryBackoffLinear(3)(0)).toBe(0);
  // interval 3s - retry 1
  expect(RetryBackoffLinear(3)(1)).toBe(3000);
  // interval 3s - retry 2
  expect(RetryBackoffLinear(3)(2)).toBe(6000);
  // interval 3s - retry 3
  expect(RetryBackoffLinear(3)(3)).toBe(9000);
  // interval 3s - retry 4
  expect(RetryBackoffLinear(3)(4)).toBe(12000);
});

test("retry delay - constant", () => {
  // delay 5s - retry 0
  expect(RetryConstant(5)(0)).toBe(0);
  // delay 5s - retry 1
  expect(RetryConstant(5)(1)).toBe(5000);
  // delay 5s - retry 2
  expect(RetryConstant(5)(2)).toBe(5000);
  // delay 5s - retry 3
  expect(RetryConstant(5)(3)).toBe(5000);
  // delay 5s - retry 4
  expect(RetryConstant(5)(4)).toBe(5000);
});

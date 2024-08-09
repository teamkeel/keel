import { models, actions, resetDatabase } from "@teamkeel/testing";
import { test, beforeEach, expect } from "vitest";

beforeEach(resetDatabase);

class CustomError extends Error {
  code: string;
  constructor(code: string, message: string) {
    super(message);
    this.code = code;
  }
}

test("Not found errors", async () => {
  await expect(
    (async () => {
      await actions.hookNotFound({
        id: "123",
      });
    })()
  ).rejects.toThrowError(
    new CustomError("ERR_RECORD_NOT_FOUND", "record not found")
  );

  await expect(
    (async () => {
      await actions.hookNotFoundCustomMessage({
        id: "123",
      });
    })()
  ).rejects.toThrowError(
    new CustomError("ERR_RECORD_NOT_FOUND", "nothing here")
  );
});

test("Bad request errors", async () => {
  await expect(
    (async () => {
      await actions.badRequest({
        id: "123",
      });
    })()
  ).rejects.toThrowError(
    new CustomError("ERR_INVALID_INPUT", "invalid inputs")
  );
});

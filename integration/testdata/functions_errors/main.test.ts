import { models, actions, resetDatabase } from "@teamkeel/testing";
import { test, beforeEach, expect } from "vitest";

beforeEach(resetDatabase);

test("Not found errors", async () => {
  await expect(
    (async () => {
      await actions.hookNotFound({
        id: "123",
      });
    })()
  ).rejects.toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });

  await expect(
    (async () => {
      await actions.hookNotFoundCustomMessage({
        id: "123",
      });
    })()
  ).rejects.toEqual({
    code: "ERR_RECORD_NOT_FOUND",
    message: "nothing here",
  });
});

test("Bad request errors", async () => {
  await expect(
    (async () => {
      await actions.badRequest({
        id: "123",
      });
    })()
  ).rejects.toEqual({
    code: "ERR_INVALID_INPUT",
    message: "invalid inputs",
  });
});

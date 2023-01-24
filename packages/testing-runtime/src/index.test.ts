import { expect, test } from "vitest";
import "./index";

test("toHaveAuthorizationError", async () => {
  const p = Promise.reject({
    code: "ERR_PERMISSION_DENIED",
  });
  await expect(p).toHaveAuthorizationError();
});

test("not.toHaveAuthorizationError", async () => {
  const p = Promise.resolve({
    id: "foo",
  });
  await expect(p).not.toHaveAuthorizationError();
});

test("toHaveError", async () => {
  const p = Promise.reject({
    code: "ERR_INVALID_INPUT",
    message: "Invalid input",
  });

  await expect(p).toHaveError({
    code: "ERR_INVALID_INPUT",
  });
});

test("not.toHaveError", async () => {
  const p = Promise.resolve({
    id: "foo",
  });

  await expect(p).not.toHaveError({
    code: "ERR_INVALID_INPUT",
  });
});

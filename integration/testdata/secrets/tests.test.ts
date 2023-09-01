import { actions, resetDatabase } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

beforeEach(resetDatabase);

test("create person with secret key", async () => {
  const person = await actions.createPerson({
    name: "dave",
    secretKey: "1232132_2323",
  });
});

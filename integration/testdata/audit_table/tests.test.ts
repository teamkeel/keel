import { test, expect, beforeEach } from "vitest";
import { actions, models, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase)

test("audit", async () => {
    const identity = await models.identity.create({ email: "dave.new@keel.xyz" });

   
    await  actions.withIdentity(identity).createAccount({ name: "Cheque" })
  
});
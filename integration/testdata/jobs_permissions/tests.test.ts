import { actions, models, jobs, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";


beforeEach(resetDatabase);

test("job - without identity - auth failed", async () => { 
    await expect(
         jobs.myJob({ name: "George"})
    ).toHaveAuthorizationError();
});

test("job - wrong domain - auth failed", async () => { 
    const identity = await models.identity.create({ email: "weave@gmail.com" })

    await expect(
         jobs.withIdentity(identity).myJob({ name: "George"})
    ).toHaveAuthorizationError();
});

test("job - authorised domain - no auth failure", async () => { 
    const identity = await models.identity.create({ email: "keel@keel.so" })

    await expect(
         jobs.withIdentity(identity).myJob({ name: "George"})
    ).not.toHaveAuthorizationError();
});
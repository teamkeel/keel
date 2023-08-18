import { test, expect } from "vitest";
import { subscribers, models } from "@teamkeel/testing";

test("subscriber - mutating field", async () => {
  const mary = await models.member.create({
    name: "Mary",
    email: "mary@keel.so",
  });

  const payload = {
    name: "member.create",
    model: "Member",
    sourceId: mary.id,
    occurredAt: new Date(),
    data: mary,
  };

  await subscribers.verifyEmail(payload);

  const updatedMary = await models.member.findOne({ id: mary.id });

  expect(mary?.verified).toBeFalsy();
  expect(updatedMary?.verified).toBeTruthy();
});

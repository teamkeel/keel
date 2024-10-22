import { test, expect, beforeEach } from "vitest";
import { subscribers, models, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase);

async function subscriberRan() {
  const trackers = await models.trackSubscriber.findMany();
  expect(trackers).toHaveLength(1);
  return trackers[0].didSubscriberRun;
}

test("subscriber - mutating field", async () => {
  const mary = await models.member.create({
    name: "Mary",
    email: "mary@keel.so",
  });

  const event = {
    eventName: "member.created" as const,
    occurredAt: new Date(),
    target: {
      id: mary.id,
      type: "Member",
      data: mary,
    },
  };

  await subscribers.verifyEmail(event);

  const updatedMary = await models.member.findOne({ id: event.target.data.id });

  expect(event.target.data?.verified).toBeFalsy();
  expect(updatedMary?.verified).toBeTruthy();
});

test("subscriber - exception - internal error with default rollback transaction", async () => {
  await models.trackSubscriber.create({ didSubscriberRun: false });

  const mary = await models.member.create({
    name: "Mary",
    email: "mary@keel.so",
  });

  const event = {
    eventName: "member.created" as const,
    occurredAt: new Date(),
    target: {
      id: mary.id,
      type: "Member",
      data: mary,
    },
  };

  await expect(subscribers.subscriberWithException(event)).toHaveError({
    code: "ERR_UNKNOWN",
    message: "something bad has happened!",
  });

  // This would be true if the transaction didn't roll back.
  expect(await subscriberRan()).toBeFalsy();
});

test("subscriber - with env vars - successful", async () => {
  await models.trackSubscriber.create({ didSubscriberRun: false });

  const mary = await models.member.create({
    name: "Mary",
    email: "mary@keel.so",
  });

  const event = {
    eventName: "member.created" as const,
    occurredAt: new Date(),
    target: {
      id: mary.id,
      type: "Member",
      data: mary,
    },
  };

  await expect(subscribers.subscriberEnvvars(event)).not.toHaveError({});

  expect(await subscriberRan()).toBeTruthy();
});

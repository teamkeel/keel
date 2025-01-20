import { actions, resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";
import { useDatabase, Duration } from "@teamkeel/sdk";
import { sql } from "kysely";

beforeEach(resetDatabase);

test("duration - create action with duration input", async () => {
  const result = await actions.createDuration({
    dur: Duration.fromISOString("PT2H3M4S"),
  });

  expect(result.dur).toEqual("PT2H3M4S");
});

test("duration - update action with duration input", async () => {
  const result = await actions.createDuration({
    dur: Duration.fromISOString("PT2H3M4S"),
  });

  const updated = await actions.updateDuration({
    where: {
      id: result.id,
    },
    values: {
      dur: Duration.fromISOString("PT1S"),
    },
  });

  expect(updated.dur).toEqual("PT1S");
});

test("duration - write custom function", async () => {
  const result = await actions.writeCustomFunction({
    dur: Duration.fromISOString("PT1H2M3S"),
  });

  expect(result.model.dur).toEqual("PT1H2M3S");

  const mydurs = await useDatabase()
    .selectFrom("my_duration")
    .selectAll()
    .execute();

  expect(mydurs.length).toEqual(1);
  expect(mydurs[0].id).toEqual(result.model.id);
  expect(mydurs[0].dur?.toISOString()).toEqual("PT1H2M3S");
});

test("durs - create and store duration in hook", async () => {
  await actions.createDurationInHook({});

  const mydurs = await useDatabase()
    .selectFrom("my_duration")
    .selectAll()
    .execute();

  expect(mydurs.length).toEqual(1);
  expect(mydurs[0].dur?.toISOString()).toEqual("PT1H");
});

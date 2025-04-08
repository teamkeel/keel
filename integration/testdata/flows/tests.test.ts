import { resetDatabase, models, flows } from "@teamkeel/testing";
import { useDatabase } from "@teamkeel/sdk";
import { beforeEach, expect, test } from "vitest";
import { sql } from "kysely";

beforeEach(resetDatabase);

test("flows - basic execution", async () => {
  await flows.myFlow({ name: "Keelson", age: 25 });

  const things = await models.thing.findMany();
  expect(things.length).toBe(1);
  expect(things[0].name).toBe("Keelson");

  const dbFlows = await sql`SELECT * FROM keel_flow_run`.execute(useDatabase());
  expect(dbFlows.rows.length).toBe(1);
  const dbSteps = await sql`SELECT * FROM keel_flow_step`.execute(
    useDatabase()
  );
  expect(dbSteps.rows.length).toBe(2);
});

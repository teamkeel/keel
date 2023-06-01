import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("eqAndNotEq", async () => {
  await models.thing.create({ switchIsOn: null });
  await models.thing.create({ switchIsOn: false });
  await models.thing.create({ switchIsOn: true });

  let resp = await actions.eqArg({ where: { arg: null } });
  expect(resp.results.map((thing) => thing.switchIsOn)).toEqual([null]);

  resp = await actions.eqArg({ where: { arg: false } });
  expect(resp.results.map((thing) => thing.switchIsOn)).toEqual([false]);

  resp = await actions.eqArg({ where: { arg: true } });
  expect(resp.results.map((thing) => thing.switchIsOn)).toEqual([true]);

  let nullsLast = function (a, b) {
    if (a === null) {
      return 1;
    }
    if (b === null) {
      return -1;
    }
    return a < b ? -1 : 1;
  };

  resp = await actions.notEqArg({ where: { arg: null } });
  expect(resp.results.map((thing) => thing.switchIsOn).sort(nullsLast)).toEqual(
    [false, true]
  );

  resp = await actions.notEqArg({ where: { arg: false } });
  expect(resp.results.map((thing) => thing.switchIsOn).sort(nullsLast)).toEqual(
    [true, null]
  );

  resp = await actions.notEqArg({ where: { arg: true } });
  expect(resp.results.map((thing) => thing.switchIsOn).sort(nullsLast)).toEqual(
    [false, null]
  );
});

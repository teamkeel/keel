import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase, actions } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("computed fields - circular reference", async () => {
  const account = await models.account.create({ standardTransactionFee: 10 });
  const transaction1 = await models.transaction.create({
    accountId: account.id,
    amount: 100,
  });
  const transaction2 = await models.transaction.create({
    accountId: account.id,
    amount: 200,
  });

  const getAccount = await models.account.findOne({ id: account.id });
  expect(getAccount!.balance).toBe(320);
});

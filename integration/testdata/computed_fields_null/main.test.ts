import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase, actions } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("computed fields - nulls in 1:M", async () => {
  const account = await models.account.create({
    fee: null,
  });

  const transaction1 = await models.transaction.create({
    accountId: account.id,
    amount: 100,
  });

  const transaction2 = await models.transaction.create({
    accountId: account.id,
    amount: null,
  });

  const transaction3 = await models.transaction.create({
    accountId: account.id,
    amount: 300,
  });

  const getAccount = await models.account.findOne({ id: account.id });
  expect(getAccount?.totalAmount).toBe(400);
  expect(getAccount?.smallestAmount).toBe(100);
  expect(getAccount?.largestAmount).toBe(300);
  expect(getAccount?.totalWithFee).toBeNull();
});

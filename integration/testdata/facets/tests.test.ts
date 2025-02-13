import { actions, resetDatabase, models } from "@teamkeel/testing";
import { beforeAll, expect, test } from "vitest";
import { Status } from "@teamkeel/sdk";


beforeAll(async () => {
  await models.order.create({
    quantity: 10,
    price: 100,
    category: "Toys",
    status: Status.Complete,
  });

  await models.order.create({
    quantity: 10,
    price: 100,
    category: "Toys",
    status: Status.InProgress,
  });

  await models.order.create({
    quantity: 1,
    price: 4100,
    category: "Computers",
    status: Status.Complete,
  });

  await models.order.create({
    quantity: 8,
    price: 80,
    category: "Toys",
    status: Status.Cancelled,
  });

  await models.order.create({
    quantity: 2,
    price: 75,
    category: "Pet Care 101",
    status: Status.InProgress,
  });

  await models.order.create({
    quantity: 5,
    price: 155,
    category: "Pet Care 101",
    status: Status.Complete,
  });
});

test("facets - no input filters", async () => {
  const result = await actions.listOrders();

  expect(result.resultInfo.quantity.min).toEqual(1);
  expect(result.resultInfo.quantity.max).toEqual(10);
  expect(result.resultInfo.quantity.avg).toEqual(5.6);

  expect(result.resultInfo.price.min).toEqual(75);
  expect(result.resultInfo.price.max).toEqual(4100);
  expect(result.resultInfo.price.avg).toEqual(906);

  expect(result.resultInfo.status["InProgress"]).toEqual(2);
  expect(result.resultInfo.status["Complete"]).toEqual(3);
  expect(result.resultInfo.status["Cancelled"]).toBeUndefined();

  expect(result.resultInfo.category["Toys"]).toEqual(2);
  expect(result.resultInfo.category["Computers"]).toEqual(1);
  expect(result.resultInfo.category["Pet Care 101"]).toEqual(2);
});

test("facets - no input filters with paging", async () => {
  const result = await actions.listOrders({
    first: 2,
  });

  expect(result.resultInfo.quantity.min).toEqual(1);
  expect(result.resultInfo.quantity.max).toEqual(10);
  expect(result.resultInfo.quantity.avg).toEqual(5.6);

  expect(result.resultInfo.price.min).toEqual(75);
  expect(result.resultInfo.price.max).toEqual(4100);
  expect(result.resultInfo.price.avg).toEqual(906);

  expect(result.resultInfo.status["InProgress"]).toEqual(2);
  expect(result.resultInfo.status["Complete"]).toEqual(3);
  expect(result.resultInfo.status["Cancelled"]).toBeUndefined();

  expect(result.resultInfo.category["Toys"]).toEqual(2);
  expect(result.resultInfo.category["Computers"]).toEqual(1);
  expect(result.resultInfo.category["Pet Care 101"]).toEqual(2);
});

test("facets - price filter", async () => {
  const result = await actions.listOrders({
    where: {
      price: {
        greaterThan: 150,
      },
    },
  });

  expect(result.resultInfo.quantity.min).toEqual(1);
  expect(result.resultInfo.quantity.max).toEqual(5);
  expect(result.resultInfo.quantity.avg).toEqual(3);

  expect(result.resultInfo.price.min).toEqual(75);
  expect(result.resultInfo.price.max).toEqual(4100);
  expect(result.resultInfo.price.avg).toEqual(906);

  expect(result.resultInfo.status["InProgress"]).toEqual(0);
  expect(result.resultInfo.status["Complete"]).toEqual(2);
  expect(result.resultInfo.status["Cancelled"]).toBeUndefined();

  expect(result.resultInfo.category["Toys"]).toEqual(0);
  expect(result.resultInfo.category["Computers"]).toEqual(1);
  expect(result.resultInfo.category["Pet Care 101"]).toEqual(1);
});
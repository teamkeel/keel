import { actions, models } from "@teamkeel/testing";
import { beforeAll, expect, test } from "vitest";
import { Status } from "@teamkeel/sdk";

beforeAll(async () => {
  await models.order.create({
    quantity: 10,
    price: 100,
    category: "Toys",
    status: Status.Complete,
    orderDate: new Date("2024-01-01"),
    orderTime: new Date("2024-01-01T12:00:00Z"),
  });

  await models.order.create({
    quantity: 10,
    price: 100,
    category: "Toys",
    status: Status.InProgress,
    orderDate: new Date("2024-01-02"),
    orderTime: new Date("2024-01-02T12:00:00Z"),
  });

  await models.order.create({
    quantity: 1,
    price: 4100,
    category: "Computers",
    status: Status.Complete,
    orderDate: new Date("2024-01-03"),
    orderTime: new Date("2024-01-03T12:00:00Z"),
  });

  await models.order.create({
    quantity: 8,
    price: 80,
    category: "Toys",
    status: Status.Cancelled,
    orderDate: new Date("2024-01-04"),
  });

  await models.order.create({
    quantity: 2,
    price: 75,
    category: "Pet Care 101",
    status: Status.InProgress,
    orderDate: new Date("2024-01-05"),
  });

  await models.order.create({
    quantity: 5,
    price: 155,
    category: "Pet Care 101",
    status: Status.Complete,
    orderDate: new Date("2024-01-06"),
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

  expect(result.resultInfo.status).toEqual([
    { value: "Complete", count: 3 },
    { value: "InProgress", count: 2 },
  ]);

  expect(result.resultInfo.category).toEqual([
    { value: "Computers", count: 1 },
    { value: "Pet Care 101", count: 2 },
    { value: "Toys", count: 2 },
  ]);

  expect(result.resultInfo.orderDate.min).toEqual(
    new Date("2024-01-01 00:00:00Z")
  );
  expect(result.resultInfo.orderDate.max).toEqual(
    new Date("2024-01-06T00:00:00Z")
  );

  expect(result.resultInfo.orderTime.min).toEqual(
    new Date("2024-01-01T12:00:00Z")
  );
  expect(result.resultInfo.orderTime.max).toEqual(
    new Date("2024-01-03T12:00:00Z")
  );
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

  expect(result.resultInfo.status).toEqual([
    { value: "Complete", count: 3 },
    { value: "InProgress", count: 2 },
  ]);

  expect(result.resultInfo.category).toEqual([
    { value: "Computers", count: 1 },
    { value: "Pet Care 101", count: 2 },
    { value: "Toys", count: 2 },
  ]);

  expect(result.resultInfo.orderDate.min).toEqual(
    new Date("2024-01-01 00:00:00Z")
  );
  expect(result.resultInfo.orderDate.max).toEqual(
    new Date("2024-01-06T00:00:00Z")
  );

  expect(result.resultInfo.orderTime.min).toEqual(
    new Date("2024-01-01T12:00:00Z")
  );
  expect(result.resultInfo.orderTime.max).toEqual(
    new Date("2024-01-03T12:00:00Z")
  );
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

  expect(result.resultInfo.status).toEqual([{ value: "Complete", count: 2 }]);

  expect(result.resultInfo.category).toEqual([
    { value: "Computers", count: 1 },
    { value: "Pet Care 101", count: 1 },
  ]);

  expect(result.resultInfo.orderDate.min).toEqual(
    new Date("2024-01-03 00:00:00Z")
  );
  expect(result.resultInfo.orderDate.max).toEqual(
    new Date("2024-01-06T00:00:00Z")
  );

  expect(result.resultInfo.orderTime.min).toEqual(
    new Date("2024-01-03T12:00:00Z")
  );
  expect(result.resultInfo.orderTime.max).toEqual(
    new Date("2024-01-03T12:00:00Z")
  );
});

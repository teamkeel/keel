import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

test("create relationships - create 1:M and associate M:1", async () => {
  const product1 = await models.product.create({
    name: "Hair Dryer",
  });

  const product2 = await models.product.create({
    name: "Hair Clips",
  });

  const customer = await models.customer.create({
    name: "Madonna",
  });

  const order = await actions.createOrder({
    onPromotion: true,
    customer: { id: customer.id },
    items: [
      { quantity: 1, price: 25, product: { id: product1.id } },
      { quantity: 15, price: 2, product: { id: product2.id } },
    ],
  });

  const orderItems = await models.orderItem.findMany({
    order: { id: order.id },
  });

  expect(order.customerId).toEqual(customer.id);
  expect(orderItems.length).toEqual(2);

  const orderItemForProduct1 = await models.orderItem.findMany({
    order: { id: order.id },
    product: { id: product1.id },
  });
  expect(orderItemForProduct1.length).toEqual(1);
  expect(orderItemForProduct1[0].quantity).toEqual(1);
  expect(orderItemForProduct1[0].price).toEqual(25);
  expect(orderItemForProduct1[0].productId).toEqual(product1.id);

  const orderItemForProduct2 = await models.orderItem.findMany({
    order: { id: order.id },
    product: { id: product2.id },
  });
  expect(orderItemForProduct2.length).toEqual(1);
  expect(orderItemForProduct2[0].quantity).toEqual(15);
  expect(orderItemForProduct2[0].price).toEqual(2);
  expect(orderItemForProduct2[0].productId).toEqual(product2.id);
});

test("create relationships - create 1:M and create M:1", async () => {
  const order = await actions.createOrderWithRelated({
    onPromotion: true,
    customer: { name: "Madonna" },
    items: [
      { quantity: 1, price: 25, product: { name: "Hair Dryer" } },
      { quantity: 15, price: 2, product: { name: "Hair Clips" } },
    ],
  });

  const customer = await models.customer.findOne({ id: order.customerId });

  expect(customer!.id).toEqual(order.customerId);
  expect(customer!.name).toEqual("Madonna");

  const orderItems = await models.orderItem.findMany({
    order: { id: order.id },
  });
  expect(orderItems.length).toEqual(2);

  const orderItemForProduct1 = await models.orderItem.findMany({
    order: { id: order.id },
    product: { name: "Hair Dryer" },
  });
  expect(orderItemForProduct1.length).toEqual(1);
  expect(orderItemForProduct1[0].quantity).toEqual(1);
  expect(orderItemForProduct1[0].price).toEqual(25);

  const orderItemForProduct2 = await models.orderItem.findMany({
    order: { id: order.id },
    product: { name: "Hair Clips" },
  });
  expect(orderItemForProduct2.length).toEqual(1);
  expect(orderItemForProduct2[0].quantity).toEqual(15);
  expect(orderItemForProduct2[0].price).toEqual(2);
});

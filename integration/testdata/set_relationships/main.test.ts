import { actions, resetDatabase, models } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("set by relationship lookup in create action", async () => {
  const customer = await actions.createCustomer({
    discountProfile: {
      discountPercentage: 10,
    },
  });

  const product = await models.product.create({
    standardPrice: 100,
  });

  const order = await actions.createOrder({
    product: { id: product.id },
    quantity: 3,
    customer: { id: customer.id },
  });

  expect(order.price).toEqual(100);
  expect(order.discountPercentage).toEqual(10);
  expect(order.discount).toEqual(30);
  expect(order.total).toEqual(270);
});

test("set by relationship lookup in update action", async () => {
    const customer = await actions.createCustomer({
      discountProfile: {
        discountPercentage: 10,
      },
    });
  
    const product = await models.product.create({
      standardPrice: 100,
    });
  
    const order = await actions.createOrder({
      product: { id: product.id },
      quantity: 3,
      customer: { id: customer.id },
    });
  
    await models.customerDiscount.update({
        customerId: customer.id}, {
        discountPercentage: 40,
        }
    );

    await models.product.update({
        id: product.id }, {
        standardPrice: 120,
    });

    const resetOrder = await actions.resetTotal({ where: { id: order.id } });
    
    expect(resetOrder?.price).toEqual(120);
    expect(resetOrder?.discountPercentage).toEqual(40);
    expect(resetOrder?.discount).toEqual(144);
    expect(resetOrder?.total).toEqual(216);
  });
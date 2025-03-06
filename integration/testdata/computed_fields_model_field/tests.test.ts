import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("computed fields - model field", async () => {
  const customer1 = await models.customer.create({ name: "John Doe" });
  const order = await models.order.create({ customerId: customer1.id });
  const job = await models.orderJob.create({ orderId: order.id });

  expect(job.customerId).toEqual(customer1.id);

  const customer2 = await models.customer.create({ name: "Weave Keelson" });
  const updatedOrder = await models.order.update(
    { id: order.id },
    { customerId: customer2.id }
  );
  expect(updatedOrder.customerId).toEqual(customer2.id);

  const getJob = await models.orderJob.findOne({ id: job.id });
  expect(getJob?.customerId).toEqual(customer2.id);
});

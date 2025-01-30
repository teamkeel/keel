import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase, actions } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("computed fields - one to many", async () => {
  const product1 = await models.product.create({
    price: 100,
  });

  const product2 = await models.product.create({
    price: 200,
  });

  const invoiceA = await actions.createInvoice({ shipping: 5 });
  expect(invoiceA.total).toBe(5);

  const item1 = await models.item.create({
    invoiceId: invoiceA.id,
    productId: product1.id,
  });

  const item2 = await models.item.create({
    invoiceId: invoiceA.id,
    productId: product2.id,
  });

  const invoiceB = await actions.createInvoice({ shipping: 5 });
  expect(invoiceB.total).toBe(5);

  const item3 = await models.item.create({
    invoiceId: invoiceB.id,
    productId: product2.id,
  });

  const inv1A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv1A?.total).toBe(305);

  const inv1B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv1B?.total).toBe(205);

  await models.product.update({ id: product1.id }, { price: 150 });

  const inv2A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv2A?.total).toBe(355);

  const inv2B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv2B?.total).toBe(205);

  await models.item.delete({ id: item2.id });

  const inv3A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv3A?.total).toBe(155);

  const inv3B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv3B?.total).toBe(205);

  const item4 = await models.item.create({
    invoiceId: invoiceA.id,
    productId: product2.id,
  });

  const inv4A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv4A?.total).toBe(355);

  const inv4B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv4B?.total).toBe(205);

  await models.item.update({ id: item4.id }, { invoiceId: invoiceB.id });

  const inv5A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv5A?.total).toBe(155);

  const inv5B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv5B?.total).toBe(405);

  await models.product.delete({ id: product2.id });

  const inv6A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv6A?.total).toBe(155);

  const inv6B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv6B?.total).toBe(5);

  await models.invoice.update({ id: invoiceA.id }, { shipping: 10 });

  const inv7A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv7A?.total).toBe(160);
});

test("computed fields - one to many - nested create", async () => {
  const product1 = await models.product.create({
    price: 100,
  });

  const product2 = await models.product.create({
    price: 200,
  });

  const invoice = await actions.createInvoice({
    shipping: 5,
    items: [
      {
        product: { id: product1.id },
      },
      {
        product: { id: product2.id },
      },
    ],
  });

  expect(invoice.total).toBe(305);
});

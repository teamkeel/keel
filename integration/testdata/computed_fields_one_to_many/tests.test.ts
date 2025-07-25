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

  await models.product.create({
    price: 180,
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
  expect(inv1A?.totalActive).toBe(305);
  expect(inv1A?.numberOfItems).toBe(2);
  expect(inv1A?.numberOfActiveItems).toBe(2);

  const inv1B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv1B?.total).toBe(205);
  expect(inv1B?.totalActive).toBe(205);
  expect(inv1B?.numberOfItems).toBe(1);
  expect(inv1B?.numberOfActiveItems).toBe(1);

  await models.product.update({ id: product1.id }, { price: 150 });

  const inv2A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv2A?.total).toBe(355);
  expect(inv2A?.totalActive).toBe(355);
  expect(inv2A?.numberOfItems).toBe(2);
  expect(inv2A?.numberOfActiveItems).toBe(2);

  const inv2B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv2B?.total).toBe(205);
  expect(inv2B?.totalActive).toBe(205);
  expect(inv2B?.numberOfItems).toBe(1);
  expect(inv2B?.numberOfActiveItems).toBe(1);

  await models.item.update({ id: item2.id }, { isDeleted: true });

  const inv22A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv22A?.total).toBe(355);
  expect(inv22A?.totalActive).toBe(155);
  expect(inv22A?.numberOfItems).toBe(2);
  expect(inv22A?.numberOfActiveItems).toBe(1);

  const inv22B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv22B?.total).toBe(205);
  expect(inv22B?.totalActive).toBe(205);
  expect(inv22B?.numberOfItems).toBe(1);
  expect(inv22B?.numberOfActiveItems).toBe(1);

  await models.item.update({ id: item2.id }, { isDeleted: false });

  const inv22A2 = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv22A2?.total).toBe(355);
  expect(inv22A2?.totalActive).toBe(355);
  expect(inv22A2?.numberOfItems).toBe(2);
  expect(inv22A2?.numberOfActiveItems).toBe(2);

  const inv22B2 = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv22B2?.total).toBe(205);
  expect(inv22B2?.totalActive).toBe(205);
  expect(inv22B2?.numberOfItems).toBe(1);
  expect(inv22B2?.numberOfActiveItems).toBe(1);

  await models.item.delete({ id: item2.id });

  const inv3A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv3A?.total).toBe(155);
  expect(inv3A?.totalActive).toBe(155);
  expect(inv3A?.numberOfItems).toBe(1);
  expect(inv3A?.numberOfActiveItems).toBe(1);

  const inv3B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv3B?.total).toBe(205);
  expect(inv3B?.totalActive).toBe(205);
  expect(inv3B?.numberOfItems).toBe(1);
  expect(inv3B?.numberOfActiveItems).toBe(1);

  const item4 = await models.item.create({
    invoiceId: invoiceA.id,
    productId: product2.id,
  });

  const inv4A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv4A?.total).toBe(355);
  expect(inv4A?.totalActive).toBe(355);
  expect(inv4A?.numberOfItems).toBe(2);
  expect(inv4A?.numberOfActiveItems).toBe(2);

  const inv4B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv4B?.total).toBe(205);
  expect(inv4B?.totalActive).toBe(205);
  expect(inv4B?.numberOfItems).toBe(1);
  expect(inv4B?.numberOfActiveItems).toBe(1);

  await models.item.update({ id: item4.id }, { invoiceId: invoiceB.id });

  const inv5A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv5A?.total).toBe(155);
  expect(inv5A?.totalActive).toBe(155);
  expect(inv5A?.numberOfItems).toBe(1);
  expect(inv5A?.numberOfActiveItems).toBe(1);

  const inv5B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv5B?.total).toBe(405);
  expect(inv5B?.totalActive).toBe(405);
  expect(inv5B?.numberOfItems).toBe(2);
  expect(inv5B?.numberOfActiveItems).toBe(2);

  await models.product.delete({ id: product2.id });

  const inv6A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv6A?.total).toBe(155);
  expect(inv6A?.totalActive).toBe(155);
  expect(inv6A?.numberOfItems).toBe(1);
  expect(inv6A?.numberOfActiveItems).toBe(1);

  const inv6B = await models.invoice.findOne({ id: invoiceB.id });
  expect(inv6B?.total).toBe(5);
  expect(inv6B?.totalActive).toBe(5);
  expect(inv6B?.numberOfItems).toBe(0);
  expect(inv6B?.numberOfActiveItems).toBe(0);

  await models.invoice.update({ id: invoiceA.id }, { shipping: 10 });

  const inv7A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv7A?.total).toBe(160);
  expect(inv7A?.totalActive).toBe(160);
  expect(inv7A?.numberOfItems).toBe(1);
  expect(inv7A?.numberOfActiveItems).toBe(1);

  await models.product.update({ id: product1.id }, { price: 0 });

  const inv8A = await models.invoice.findOne({ id: invoiceA.id });
  expect(inv8A?.total).toBe(0);
  expect(inv8A?.totalActive).toBe(0);
  expect(inv8A?.numberOfItems).toBe(1);
  expect(inv8A?.numberOfActiveItems).toBe(0);
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

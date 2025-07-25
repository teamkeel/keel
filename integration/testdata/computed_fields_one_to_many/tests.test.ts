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

  let invoiceA: any = await actions.createInvoice({ shipping: 5 });
  expect(invoiceA.total).toBe(5);

  const item1 = await models.item.create({
    invoiceId: invoiceA.id,
    productId: product1.id,
  });

  const item2 = await models.item.create({
    invoiceId: invoiceA.id,
    productId: product2.id,
  });

  let invoiceB: any = await actions.createInvoice({ shipping: 5 });
  expect(invoiceB.total).toBe(5);

  const item3 = await models.item.create({
    invoiceId: invoiceB.id,
    productId: product2.id,
  });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(305);
  expect(invoiceA?.totalActive).toBe(305);
  expect(invoiceA?.numberOfItems).toBe(2);
  expect(invoiceA?.numberOfActiveItems).toBe(2);
  expect(invoiceA?.hasItems).toBe(true);

  invoiceB = await models.invoice.findOne({ id: invoiceB.id });
  expect(invoiceB?.total).toBe(205);
  expect(invoiceB?.totalActive).toBe(205);
  expect(invoiceB?.numberOfItems).toBe(1);
  expect(invoiceB?.numberOfActiveItems).toBe(1);
  expect(invoiceB?.hasItems).toBe(true);

  await models.product.update({ id: product1.id }, { price: 150 });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(355);
  expect(invoiceA?.totalActive).toBe(355);
  expect(invoiceA?.numberOfItems).toBe(2);
  expect(invoiceA?.numberOfActiveItems).toBe(2);
  expect(invoiceA?.hasItems).toBe(true);

  invoiceB = await models.invoice.findOne({ id: invoiceB.id });
  expect(invoiceB?.total).toBe(205);
  expect(invoiceB?.totalActive).toBe(205);
  expect(invoiceB?.numberOfItems).toBe(1);
  expect(invoiceB?.numberOfActiveItems).toBe(1);
  expect(invoiceB?.hasItems).toBe(true);

  await models.item.update({ id: item2.id }, { isDeleted: true });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(355);
  expect(invoiceA?.totalActive).toBe(155);
  expect(invoiceA?.numberOfItems).toBe(2);
  expect(invoiceA?.numberOfActiveItems).toBe(1);
  expect(invoiceA?.hasItems).toBe(true);

  invoiceB = await models.invoice.findOne({ id: invoiceB.id });
  expect(invoiceB?.total).toBe(205);
  expect(invoiceB?.totalActive).toBe(205);
  expect(invoiceB?.numberOfItems).toBe(1);
  expect(invoiceB?.numberOfActiveItems).toBe(1);
  expect(invoiceB?.hasItems).toBe(true);

  await models.item.update({ id: item2.id }, { isDeleted: false });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(355);
  expect(invoiceA?.totalActive).toBe(355);
  expect(invoiceA?.numberOfItems).toBe(2);
  expect(invoiceA?.numberOfActiveItems).toBe(2);
  expect(invoiceA?.hasItems).toBe(true);

  invoiceB = await models.invoice.findOne({ id: invoiceB.id });
  expect(invoiceB?.total).toBe(205);
  expect(invoiceB?.totalActive).toBe(205);
  expect(invoiceB?.numberOfItems).toBe(1);
  expect(invoiceB?.numberOfActiveItems).toBe(1);
  expect(invoiceB?.hasItems).toBe(true);

  await models.item.delete({ id: item2.id });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(155);
  expect(invoiceA?.totalActive).toBe(155);
  expect(invoiceA?.numberOfItems).toBe(1);
  expect(invoiceA?.numberOfActiveItems).toBe(1);
  expect(invoiceA?.hasItems).toBe(true);

  invoiceB = await models.invoice.findOne({ id: invoiceB.id });
  expect(invoiceB?.total).toBe(205);
  expect(invoiceB?.totalActive).toBe(205);
  expect(invoiceB?.numberOfItems).toBe(1);
  expect(invoiceB?.numberOfActiveItems).toBe(1);
  expect(invoiceB?.hasItems).toBe(true);

  const item4 = await models.item.create({
    invoiceId: invoiceA.id,
    productId: product2.id,
  });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(355);
  expect(invoiceA?.totalActive).toBe(355);
  expect(invoiceA?.numberOfItems).toBe(2);
  expect(invoiceA?.numberOfActiveItems).toBe(2);
  expect(invoiceA?.hasItems).toBe(true);

  invoiceB = await models.invoice.findOne({ id: invoiceB.id });
  expect(invoiceB?.total).toBe(205);
  expect(invoiceB?.totalActive).toBe(205);
  expect(invoiceB?.numberOfItems).toBe(1);
  expect(invoiceB?.numberOfActiveItems).toBe(1);
  expect(invoiceB?.hasItems).toBe(true);

  await models.item.update({ id: item4.id }, { invoiceId: invoiceB.id });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(155);
  expect(invoiceA?.totalActive).toBe(155);
  expect(invoiceA?.numberOfItems).toBe(1);
  expect(invoiceA?.numberOfActiveItems).toBe(1);
  expect(invoiceA?.hasItems).toBe(true);

  invoiceB = await models.invoice.findOne({ id: invoiceB.id });
  expect(invoiceB?.total).toBe(405);
  expect(invoiceB?.totalActive).toBe(405);
  expect(invoiceB?.numberOfItems).toBe(2);
  expect(invoiceB?.numberOfActiveItems).toBe(2);
  expect(invoiceB?.hasItems).toBe(true);

  await models.product.delete({ id: product2.id });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(155);
  expect(invoiceA?.totalActive).toBe(155);
  expect(invoiceA?.numberOfItems).toBe(1);
  expect(invoiceA?.numberOfActiveItems).toBe(1);
  expect(invoiceA?.hasItems).toBe(true);

  invoiceB = await models.invoice.findOne({ id: invoiceB.id });
  expect(invoiceB?.total).toBe(5);
  expect(invoiceB?.totalActive).toBe(5);
  expect(invoiceB?.numberOfItems).toBe(0);
  expect(invoiceB?.numberOfActiveItems).toBe(0);
  expect(invoiceB?.hasItems).toBe(false);

  await models.invoice.update({ id: invoiceA.id }, { shipping: 10 });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(160);
  expect(invoiceA?.totalActive).toBe(160);
  expect(invoiceA?.numberOfItems).toBe(1);
  expect(invoiceA?.numberOfActiveItems).toBe(1);
  expect(invoiceA?.hasItems).toBe(true);

  await models.product.update({ id: product1.id }, { price: 0 });

  invoiceA = await models.invoice.findOne({ id: invoiceA.id });
  expect(invoiceA?.total).toBe(10);
  expect(invoiceA?.totalActive).toBe(10);
  expect(invoiceA?.numberOfItems).toBe(1);
  expect(invoiceA?.numberOfActiveItems).toBe(0);
  expect(invoiceA?.hasItems).toBe(true);
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

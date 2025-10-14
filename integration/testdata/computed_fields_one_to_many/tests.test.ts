import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase, actions } from "@teamkeel/testing";
import { InvoiceStatus } from "@teamkeel/sdk";

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
    price: 100,
  });

  const item2 = await models.item.create({
    invoiceId: invoiceA.id,
    productId: product2.id,
    price: 200,
  });

  let invoiceB: any = await actions.createInvoice({ shipping: 5 });
  expect(invoiceB.total).toBe(5);

  const item3 = await models.item.create({
    invoiceId: invoiceB.id,
    productId: product2.id,
    price: 200,
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
    price: 200,
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

test("product computed fields - totalValueSold and totalUnitsSold", async () => {
  // Create products
  const product1 = await models.product.create({
    price: 100,
  });

  const product2 = await models.product.create({
    price: 200,
  });

  // Initially, no sales
  let p1 = await models.product.findOne({ id: product1.id });
  let p2 = await models.product.findOne({ id: product2.id });
  expect(p1?.totalValueSold).toBe(0);
  expect(p1?.totalUnitsSold).toBe(0);
  expect(p2?.totalValueSold).toBe(0);
  expect(p2?.totalUnitsSold).toBe(0);

  // Create a draft invoice with items
  const invoice1 = await actions.createInvoice({ shipping: 5 });
  await models.item.create({
    invoiceId: invoice1.id,
    productId: product1.id,
    price: 100,
    quantity: 2,
  });
  await models.item.create({
    invoiceId: invoice1.id,
    productId: product2.id,
    price: 200,
    quantity: 1,
  });

  // Draft invoices should not count towards sales
  p1 = await models.product.findOne({ id: product1.id });
  p2 = await models.product.findOne({ id: product2.id });
  expect(p1?.totalValueSold).toBe(0);
  expect(p1?.totalUnitsSold).toBe(0);
  expect(p2?.totalValueSold).toBe(0);
  expect(p2?.totalUnitsSold).toBe(0);

  // Mark invoice as paid
  await models.invoice.update(
    { id: invoice1.id },
    { status: InvoiceStatus.Paid }
  );

  // Now sales should be counted
  // invoice1.total = 100 + 200 + 5 = 305
  // Each product gets the full invoice total (305) since they each have one item in this invoice
  p1 = await models.product.findOne({ id: product1.id });
  p2 = await models.product.findOne({ id: product2.id });
  expect(p1?.totalValueSold).toBe(305); // invoice1.total
  expect(p1?.totalUnitsSold).toBe(2);
  expect(p2?.totalValueSold).toBe(305); // invoice1.total
  expect(p2?.totalUnitsSold).toBe(1);

  // Create another paid invoice
  const invoice2 = await actions.createInvoice({ shipping: 10 });
  await models.item.create({
    invoiceId: invoice2.id,
    productId: product1.id,
    price: 100,
    quantity: 3,
  });
  await models.item.create({
    invoiceId: invoice2.id,
    productId: product2.id,
    price: 200,
    quantity: 2,
  });
  await models.invoice.update(
    { id: invoice2.id },
    { status: InvoiceStatus.Paid }
  );

  // Sales should accumulate
  // invoice2.total = 100 + 200 + 10 = 310
  // product1 has items in invoice1 (305) + invoice2 (310) = 615
  // product2 has items in invoice1 (305) + invoice2 (310) = 615
  p1 = await models.product.findOne({ id: product1.id });
  p2 = await models.product.findOne({ id: product2.id });
  expect(p1?.totalValueSold).toBe(615); // invoice1.total + invoice2.total
  expect(p1?.totalUnitsSold).toBe(5); // 2 + 3
  expect(p2?.totalValueSold).toBe(615); // invoice1.total + invoice2.total
  expect(p2?.totalUnitsSold).toBe(3); // 1 + 2

  // Mark invoice as deleted
  await models.invoice.update({ id: invoice2.id }, { isDeleted: true });

  // Deleted invoices should not count
  p1 = await models.product.findOne({ id: product1.id });
  p2 = await models.product.findOne({ id: product2.id });
  expect(p1?.totalValueSold).toBe(305); // only invoice1
  expect(p1?.totalUnitsSold).toBe(2);
  expect(p2?.totalValueSold).toBe(305); // only invoice1
  expect(p2?.totalUnitsSold).toBe(1);

  // Undelete the invoice
  await models.invoice.update({ id: invoice2.id }, { isDeleted: false });

  // Sales should be counted again
  p1 = await models.product.findOne({ id: product1.id });
  p2 = await models.product.findOne({ id: product2.id });
  expect(p1?.totalValueSold).toBe(615); // invoice1 + invoice2
  expect(p1?.totalUnitsSold).toBe(5);
  expect(p2?.totalValueSold).toBe(615); // invoice1 + invoice2
  expect(p2?.totalUnitsSold).toBe(3);

  // Change invoice back to draft
  await models.invoice.update(
    { id: invoice2.id },
    { status: InvoiceStatus.Draft }
  );

  // Draft invoices should not count
  p1 = await models.product.findOne({ id: product1.id });
  p2 = await models.product.findOne({ id: product2.id });
  expect(p1?.totalValueSold).toBe(305); // only invoice1
  expect(p1?.totalUnitsSold).toBe(2);
  expect(p2?.totalValueSold).toBe(305); // only invoice1
  expect(p2?.totalUnitsSold).toBe(1);

  // Create a paid invoice but mark item as deleted
  const invoice3 = await actions.createInvoice({ shipping: 5 });
  const item = await models.item.create({
    invoiceId: invoice3.id,
    productId: product1.id,
    price: 100,
    quantity: 5,
  });
  await models.invoice.update(
    { id: invoice3.id },
    { status: InvoiceStatus.Paid }
  );

  // invoice3.total = 100 + 5 = 105
  // product1 has items in invoice1 (305) + invoice3 (105) = 410
  p1 = await models.product.findOne({ id: product1.id });
  expect(p1?.totalValueSold).toBe(410); // invoice1.total + invoice3.total
  expect(p1?.totalUnitsSold).toBe(7); // 2 + 5

  // Mark item as deleted (not the invoice)
  await models.item.update({ id: item.id }, { isDeleted: true });

  // Deleted items should not count
  p1 = await models.product.findOne({ id: product1.id });
  expect(p1?.totalValueSold).toBe(305); // only invoice1
  expect(p1?.totalUnitsSold).toBe(2);

  // Undelete the item
  await models.item.update({ id: item.id }, { isDeleted: false });

  p1 = await models.product.findOne({ id: product1.id });
  expect(p1?.totalValueSold).toBe(410); // invoice1 + invoice3
  expect(p1?.totalUnitsSold).toBe(7);
});

import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase } from "@teamkeel/testing";
import { OrderStatus, PaymentStatus } from "@teamkeel/sdk";

beforeEach(resetDatabase);

test("computed fields - decimal", async () => {
  const item = await models.computedDecimal.create({ price: 5, quantity: 2 });
  expect(item.total).toEqual(10);
  expect(item.totalWithShipping).toEqual(15);
  expect(item.totalWithDiscount).toEqual(9);

  const get = await models.computedDecimal.findOne({ id: item.id });
  expect(get!.total).toEqual(10);
  expect(get!.totalWithShipping).toEqual(15);
  expect(get!.totalWithDiscount).toEqual(9);

  const updatePrice = await models.computedDecimal.update(
    { id: item.id },
    { price: 10 }
  );
  expect(updatePrice.total).toEqual(20);
  expect(updatePrice.totalWithShipping).toEqual(25);
  expect(updatePrice.totalWithDiscount).toEqual(18);

  const updateQuantity = await models.computedDecimal.update(
    { id: item.id },
    { quantity: 3 }
  );
  expect(updateQuantity.total).toEqual(30);
  expect(updateQuantity.totalWithShipping).toEqual(35);
  expect(updateQuantity.totalWithDiscount).toEqual(27);

  const updateBoth = await models.computedDecimal.update(
    { id: item.id },
    { price: 12, quantity: 4 }
  );
  expect(updateBoth.total).toEqual(48);
  expect(updateBoth.totalWithShipping).toEqual(53);
  expect(updateBoth.totalWithDiscount).toEqual(43.2);
});

test("computed fields - number", async () => {
  const item = await models.computedNumber.create({ price: 5, quantity: 2 });
  expect(item.total).toEqual(10);
  expect(item.totalWithShipping).toEqual(15);
  expect(item.totalWithDiscount).toEqual(9);

  const get = await models.computedNumber.findOne({ id: item.id });
  expect(get!.total).toEqual(10);
  expect(get!.totalWithShipping).toEqual(15);
  expect(get!.totalWithDiscount).toEqual(9);

  const updatePrice = await models.computedNumber.update(
    { id: item.id },
    { price: 10 }
  );
  expect(updatePrice.total).toEqual(20);
  expect(updatePrice.totalWithShipping).toEqual(25);
  expect(updatePrice.totalWithDiscount).toEqual(18);

  const updateQuantity = await models.computedNumber.update(
    { id: item.id },
    { quantity: 3 }
  );
  expect(updateQuantity.total).toEqual(30);
  expect(updateQuantity.totalWithShipping).toEqual(35);
  expect(updateQuantity.totalWithDiscount).toEqual(27);

  const updateBoth = await models.computedNumber.update(
    { id: item.id },
    { price: 12, quantity: 4 }
  );
  expect(updateBoth.total).toEqual(48);
  expect(updateBoth.totalWithShipping).toEqual(53);
  expect(updateBoth.totalWithDiscount).toEqual(43);
});

test("computed fields - boolean", async () => {
  const expensive = await models.computedBool.create({
    price: 200,
    isActive: true,
  });
  expect(expensive.isExpensive).toBeTruthy();
  expect(expensive.isCheap).toBeFalsy();

  const notExpensive = await models.computedBool.create({
    price: 90,
    isActive: true,
  });
  expect(notExpensive.isExpensive).toBeFalsy();
  expect(notExpensive.isCheap).toBeTruthy();

  const notActive = await models.computedBool.create({
    price: 200,
    isActive: false,
  });
  expect(notActive.isExpensive).toBeFalsy();
  expect(notActive.isCheap).toBeTruthy();
});

test("computed fields - text", async () => {
  const john = await models.computedText.create({
    firstName: "John",
    lastName: "Doe",
  });
  expect(john.displayName).toBe("John Doe");
  expect(john.fullDisplayName).toBe("Product: John Doe");
});

test("computed fields - with nulls", async () => {
  const item = await models.computedNulls.create({ price: 5 });
  expect(item.total).toBeNull();

  const updateQty = await models.computedNulls.update(
    { id: item.id },
    { quantity: 10 }
  );
  expect(updateQty!.total).toEqual(50);

  const updatePrice2 = await models.computedNulls.update(
    { id: item.id },
    { price: null }
  );
  expect(updatePrice2!.total).toBeNull();
});

test("computed fields - with dependencies", async () => {
  const item = await models.computedDepends.create({ price: 5, quantity: 2 });
  expect(item.total).toEqual(10);
  expect(item.totalWithShipping).toEqual(15);
  expect(item.totalWithDiscount).toEqual(14);

  const updatedQty = await models.computedDepends.update(
    { id: item.id },
    { quantity: 11 }
  );
  expect(updatedQty.total).toEqual(55);
  expect(updatedQty.totalWithShipping).toEqual(60);
  expect(updatedQty.totalWithDiscount).toEqual(54.5);

  const updatePrice = await models.computedDepends.update(
    { id: item.id },
    { price: 8 }
  );
  expect(updatePrice.total).toEqual(88);
  expect(updatePrice.totalWithShipping).toEqual(93);
  expect(updatePrice.totalWithDiscount).toEqual(84.2);
});

test("computed fields - enums", async () => {
  const item = await models.computedEnums.create({
    orderStatus: OrderStatus.NEW,
    paymentStatus: PaymentStatus.PAID,
  });
  expect(item.isComplete).toBeFalsy();

  const updated = await models.computedEnums.update(
    { id: item.id },
    { orderStatus: OrderStatus.SHIPPED, paymentStatus: PaymentStatus.UNPAID }
  );
  expect(updated.isComplete).toBeFalsy();

  const updateComplete = await models.computedEnums.update(
    { id: item.id },
    { orderStatus: OrderStatus.DELIVERED, paymentStatus: PaymentStatus.PAID }
  );
  expect(updateComplete.isComplete).toBeTruthy();
});

import { actions, models, resetDatabase } from "@teamkeel/testing";
import { Product, Customer, Order } from "@teamkeel/sdk";
import { test, describe, expect, beforeEach, beforeAll } from "vitest";

let productLaptop: Product | null;
let productMouse: Product | null;
let productKeyboard: Product | null;
let productMonitor: Product | null;

let johnDoe: Customer | null;
let pamSmith: Customer | null;

let order: Order | null;

test("purchase new products", async () => {
  productLaptop = await actions.createProduct({
    name: "Laptop",
    costPrice: 100,
    markup: 0.2,
  });

  productMouse = await actions.createProduct({
    name: "Mouse",
    costPrice: 12,
    markup: 0.4,
  });

  productKeyboard = await actions.createProduct({
    name: "Keyboard",
    costPrice: 18,
    markup: 0.4,
  });

  productMonitor = await actions.createProduct({
    name: "Monitor",
    costPrice: 50,
    markup: 0.4,
  });

  expect(productLaptop.price).toBe(120);
  expect(productMouse.price).toBe(16.8);
  expect(productKeyboard.price).toBe(25.2);
  expect(productMonitor.price).toBe(70);

  await actions.createPurchaseOrder({
    product: { id: productLaptop.id },
    quantity: 10,
  });

  await actions.createPurchaseOrder({
    product: { id: productMouse.id },
    quantity: 20,
  });

  await actions.createPurchaseOrder({
    product: { id: productKeyboard.id },
    quantity: 25,
  });

  await actions.createPurchaseOrder({
    product: { id: productMonitor.id },
    quantity: 10,
  });
});

test("check stock levels after purchase order", async () => {
  productLaptop = await actions.getProduct({ id: productLaptop!.id });
  expect(productLaptop?.stockQuantity).toBe(10);

  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(20);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(25);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(10);
});

test("create customer", async () => {
  johnDoe = await actions.createCustomer({
    name: "John Doe",
  });
});

test("check customer statistics before order", async () => {
  expect(johnDoe?.totalOrders).toBe(0);
  expect(johnDoe?.totalSpent).toBe(0);
  expect(johnDoe?.averageOrderValue).toBe(0);
  expect(johnDoe?.smallestOrder).toBe(0);
  expect(johnDoe?.largestOrder).toBe(0);
});

test("create order for new products", async () => {
  order = await actions.createOrder({
    customer: { id: johnDoe!.id },
    orderItems: [
      { product: { id: productLaptop!.id }, quantity: 2 },
      { product: { id: productMouse!.id }, quantity: 2 },
      { product: { id: productKeyboard!.id }, quantity: 1 },
    ],
  });

  expect(order.shipping).toBe(10);
  expect(order.total).toBe(240 + 33.6 + 25.2 + order.shipping);
});

test("check stock levels after order", async () => {
  productLaptop = await actions.getProduct({ id: productLaptop!.id });
  expect(productLaptop?.stockQuantity).toBe(8);

  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(18);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(24);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(10);
});

test("check customer statistics after order", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalOrders).toBe(1);
  expect(johnDoe?.totalSpent).toBe(308.8);
  expect(johnDoe?.averageOrderValue).toBe(308.8);
  expect(johnDoe?.smallestOrder).toBe(308.8);
  expect(johnDoe?.largestOrder).toBe(308.8);
});

test("adjust quantity in order", async () => {
  const items = await actions.listOrderItems({
    where: { order: { id: { equals: order?.id } } },
  });

  for (const item of items.results) {
    if (item.productId === productMouse!.id) {
      await actions.updateOrderItem({
        where: { id: item.id },
        values: { quantity: 1 },
      });
    }
  }

  order = await actions.getOrder({ id: order!.id });

  expect(order?.shipping).toBe(8);
  expect(order?.total).toBe(240 + 16.8 + 25.2 + order!.shipping);
});

test("check customer statistics after adjusting quantity", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalSpent).toBe(290);
  expect(johnDoe?.totalOrders).toBe(1);
  expect(johnDoe?.averageOrderValue).toBe(290);
  expect(johnDoe?.smallestOrder).toBe(290);
  expect(johnDoe?.largestOrder).toBe(290);
});

test("check stock levels after adjusting quantity", async () => {
  productLaptop = await actions.getProduct({ id: productLaptop!.id });
  expect(productLaptop?.stockQuantity).toBe(8);

  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(19);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(24);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(10);
});

test("change product in order item", async () => {
  const items = await actions.listOrderItems({
    where: { order: { id: { equals: order?.id } } },
  });

  for (const item of items.results) {
    if (item.productId === productMouse!.id) {
      await actions.updateOrderItem({
        where: { id: item.id },
        values: { product: { id: productMonitor!.id } },
      });
    }
  }

  order = await actions.getOrder({ id: order!.id });

  expect(order?.shipping).toBe(8);
  expect(order?.total).toBe(240 + 70 + 25.2 + order!.shipping);
});

test("check stock levels after adjusting product", async () => {
  productLaptop = await actions.getProduct({ id: productLaptop!.id });
  expect(productLaptop?.stockQuantity).toBe(8);

  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(20);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(24);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(9);
});

test("check customer statistics after adjusting product", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalSpent).toBe(343.2);
  expect(johnDoe?.totalOrders).toBe(1);
  expect(johnDoe?.averageOrderValue).toBe(343.2);
  expect(johnDoe?.smallestOrder).toBe(343.2);
  expect(johnDoe?.largestOrder).toBe(343.2);
});

test("create another order", async () => {
  order = await actions.createOrder({
    customer: { id: johnDoe!.id },
    orderItems: [{ product: { id: productMouse!.id }, quantity: 4 }],
  });

  expect(order.shipping).toBe(8);
  expect(order.total).toBe(67.2 + order.shipping);
});

test("check customer statistics after adjusting product", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalSpent).toBe(418.4);
  expect(johnDoe?.totalOrders).toBe(2);
  expect(johnDoe?.averageOrderValue).toBe(209.2);
  expect(johnDoe?.smallestOrder).toBe(75.2);
  expect(johnDoe?.largestOrder).toBe(343.2);
});

test("check stock levels after adjusting quantity", async () => {
  productLaptop = await actions.getProduct({ id: productLaptop!.id });
  expect(productLaptop?.stockQuantity).toBe(8);

  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(16);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(24);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(9);
});

test("change order's customer", async () => {
  pamSmith = await actions.createCustomer({
    name: "Pam Smith",
  });

  order = await actions.updateOrder({
    where: { id: order!.id },
    values: {
      customer: { id: pamSmith!.id },
    },
  });

  expect(order.shipping).toBe(8);
  expect(order.total).toBe(67.2 + order.shipping);
});

test("check that stock levels are the same", async () => {
  productLaptop = await actions.getProduct({ id: productLaptop!.id });
  expect(productLaptop?.stockQuantity).toBe(8);

  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(16);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(24);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(9);
});

test("check customer statistics after adjusting product", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalSpent).toBe(343.2);
  expect(johnDoe?.totalOrders).toBe(1);
  expect(johnDoe?.averageOrderValue).toBe(343.2);

  pamSmith = await actions.getCustomer({ id: pamSmith!.id });
  expect(pamSmith?.totalSpent).toBe(75.2);
  expect(pamSmith?.totalOrders).toBe(1);
  expect(pamSmith?.averageOrderValue).toBe(75.2);
});

test("fix product markup", async () => {
  productLaptop = await actions.updateProduct({
    where: { id: productLaptop!.id },
    values: { markup: 0.4 },
  });

  expect(productLaptop?.price).toBe(140);
});

test("check customer statistics after fixing product markup", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalSpent).toBe(383.2);
  expect(johnDoe?.totalOrders).toBe(1);
  expect(johnDoe?.averageOrderValue).toBe(383.2);

  pamSmith = await actions.getCustomer({ id: pamSmith!.id });
  expect(pamSmith?.totalSpent).toBe(75.2);
  expect(pamSmith?.totalOrders).toBe(1);
  expect(pamSmith?.averageOrderValue).toBe(75.2);
});

test("delete order item", async () => {
  const items = await actions.listOrderItems({
    where: { order: { id: { equals: order?.id } } },
  });
  console.log(items.results);
  for (const item of items.results) {
    if (item.productId === productMouse!.id) {
      await actions.deleteOrderItem({ id: item!.id });
    }
  }

  order = await actions.getOrder({ id: order!.id });

  expect(order?.shipping).toBe(0);
  expect(order?.total).toBe(0);
});

test("check customer statistics after fixing product markup", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalSpent).toBe(383.2);
  expect(johnDoe?.totalOrders).toBe(1);
  expect(johnDoe?.averageOrderValue).toBe(383.2);

  pamSmith = await actions.getCustomer({ id: pamSmith!.id });
  expect(pamSmith?.totalSpent).toBe(0);
  expect(pamSmith?.totalOrders).toBe(1);
  expect(pamSmith?.averageOrderValue).toBe(0);
});

test("check that stock levels have increased", async () => {
  productLaptop = await actions.getProduct({ id: productLaptop!.id });
  expect(productLaptop?.stockQuantity).toBe(8);

  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(20);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(24);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(9);
});

test("readd order item", async () => {
  await actions.addOrderItem({
    order: { id: order!.id },
    product: { id: productMouse!.id },
    quantity: 4,
  });
});

test("check customer statistics after readding order item", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalSpent).toBe(383.2);
  expect(johnDoe?.totalOrders).toBe(1);
  expect(johnDoe?.averageOrderValue).toBe(383.2);

  pamSmith = await actions.getCustomer({ id: pamSmith!.id });
  expect(pamSmith?.totalSpent).toBe(75.2);
  expect(pamSmith?.totalOrders).toBe(1);
  expect(pamSmith?.averageOrderValue).toBe(75.2);
});

test("check that stock levels after readding order item", async () => {
  productLaptop = await actions.getProduct({ id: productLaptop!.id });
  expect(productLaptop?.stockQuantity).toBe(8);

  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(16);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(24);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(9);
});

test("delete order", async () => {
  await actions.deleteOrder({ id: order!.id });
});

test("check customer statistics after deleting order", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalSpent).toBe(383.2);
  expect(johnDoe?.totalOrders).toBe(1);
  expect(johnDoe?.averageOrderValue).toBe(383.2);

  pamSmith = await actions.getCustomer({ id: pamSmith!.id });
  expect(pamSmith?.totalSpent).toBe(0);
  expect(pamSmith?.totalOrders).toBe(0);
  expect(pamSmith?.averageOrderValue).toBe(0);
  expect(pamSmith?.smallestOrder).toBe(0);
  expect(pamSmith?.largestOrder).toBe(0);
});

test("check that stock levels after deleting order", async () => {
  productLaptop = await actions.getProduct({ id: productLaptop!.id });
  expect(productLaptop?.stockQuantity).toBe(8);

  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(20);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(24);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(9);
});

test("delete product", async () => {
  await actions.deleteProduct({ id: productLaptop!.id });
});

test("check customer statistics after deleting product", async () => {
  johnDoe = await actions.getCustomer({ id: johnDoe!.id });
  expect(johnDoe?.totalSpent).toBe(99.2);
  expect(johnDoe?.totalOrders).toBe(1);
  expect(johnDoe?.averageOrderValue).toBe(99.2);
  expect(pamSmith?.smallestOrder).toBe(0); // this is because the order actually still exists
  expect(pamSmith?.largestOrder).toBe(0); //TODO?

  pamSmith = await actions.getCustomer({ id: pamSmith!.id });
  expect(pamSmith?.totalSpent).toBe(0);
  expect(pamSmith?.totalOrders).toBe(0);
  expect(pamSmith?.averageOrderValue).toBe(0);
});

test("delete customers", async () => {
  await actions.deleteCustomer({ id: johnDoe!.id });
  await actions.deleteCustomer({ id: pamSmith!.id });
});

test("check that stock levels after deleting customers", async () => {
  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(20);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(25);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(10);
});

test("delete purchase orders", async () => {
  const result = await actions.listPurchaseOrders();

  for (const purchaseOrder of result.results) {
    await actions.deletePurchaseOrder({ id: purchaseOrder!.id });
  }
});

test("check that stock levels after deleting purchase orders", async () => {
  productMouse = await actions.getProduct({ id: productMouse!.id });
  expect(productMouse?.stockQuantity).toBe(0);

  productKeyboard = await actions.getProduct({ id: productKeyboard!.id });
  expect(productKeyboard?.stockQuantity).toBe(0);

  productMonitor = await actions.getProduct({ id: productMonitor!.id });
  expect(productMonitor?.stockQuantity).toBe(0);
});

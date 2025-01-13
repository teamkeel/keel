import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase, actions } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("computed fields - many to one", async () => {
  const agent = await models.agent.create({ commission: 2.5 });
  const product = await models.product.create({ standardPrice: 5, agentId: agent.id })
  const item = await models.item.create({ 
    productId: product.id,
    quantity: 2 
  });
  expect(item.total).toEqual(12.5);
  expect(item.totalWithShipping).toEqual(17.5);
  expect(item.totalWithDiscount).toEqual(16.25);

  await models.product.update(
    { id: product.id },
    { standardPrice: 10 }
  );
  const getUpdatedPrice = await models.item.findOne({ id: item.id });
  expect(getUpdatedPrice!.total).toEqual(22.5);
  expect(getUpdatedPrice!.totalWithShipping).toEqual(27.5);
  expect(getUpdatedPrice!.totalWithDiscount).toEqual(25.25);

  const updateQuantity = await models.item.update(
    { id: item.id },
    { quantity: 3 }
  );
  expect(updateQuantity.total).toEqual(32.5);
  expect(updateQuantity.totalWithShipping).toEqual(37.5);
  expect(updateQuantity.totalWithDiscount).toEqual(34.25);

  await models.agent.update({id: agent.id}, {commission:10});
  const getUpdatedPrice2 = await models.item.findOne({ id: item.id });
  expect(getUpdatedPrice2!.total).toEqual(40);
  expect(getUpdatedPrice2!.totalWithShipping).toEqual(45);
  expect(getUpdatedPrice2!.totalWithDiscount).toEqual(41);
});

test("computed fields - many to one cascading update", async () => {
  const agent1 = await models.agent.create({ commission: 2.5 });
  const agent2 = await models.agent.create({ commission: 0.75 });

  const product1 = await models.product.create({ standardPrice: 5, agentId: agent1.id })
  const item1 = await models.item.create({ productId: product1.id,quantity: 2 });
  expect(item1.total).toEqual(12.5);

  const product2 = await models.product.create({ standardPrice: 7, agentId: agent1.id })
  const item2 = await models.item.create({ productId: product2.id,quantity: 2 });
  expect(item2.total).toEqual(16.5);

  const product3 = await models.product.create({ standardPrice: 9, agentId: agent2.id })
  const item3 = await models.item.create({ productId: product3.id,quantity: 2 });
  expect(item3.total).toEqual(18.75);

  const product4 = await models.product.create({ standardPrice: 10, agentId: agent2.id })
  const item4 = await models.item.create({ productId: product4.id,quantity: 2 });
  expect(item4.total).toEqual(20.75);
  const item5 = await models.item.create({ productId: product4.id,quantity: 3 });
  expect(item5.total).toEqual(30.75);

  await models.agent.update({id: agent2.id}, {commission:1});

  const getItem1 = await models.item.findOne({ id: item1.id });
  expect(getItem1!.total).toEqual(12.5);

  const getItem2 = await models.item.findOne({ id: item2.id });
  expect(getItem2!.total).toEqual(16.5);

  const getItem3 = await models.item.findOne({ id: item3.id });
  expect(getItem3!.total).toEqual(19);

  const getItem4 = await models.item.findOne({ id: item4.id });
  expect(getItem4!.total).toEqual(21);

  const getItem5 = await models.item.findOne({ id: item5.id });
  expect(getItem5!.total).toEqual(31);

});

test("computed fields - many to one with nested create", async () => {
  const item = await actions.createItem({
    quantity: 2,
    product: {
      standardPrice: 5,
      agent: {
        commission: 2.5
      }
    }
  });

  expect(item.total).toEqual(12.5);
});

test("computed fields - many to one with nested create from related model", async () => {
  const product = await actions.createProduct({
    standardPrice: 5,
    items:[{
      quantity: 2
    },{
      quantity: 5
    }],
    agent: {
      commission: 2.5
    }
  });

  const items = await models.item.findMany({ where: { productId: product.id }, orderBy: { total: "asc"}});

  expect(items[0].total).toEqual(12.5);
  expect(items[1].total).toEqual(27.5);
  expect(items).length(2);
});

test("computed fields - many to one with nested create from nested related model", async () => {

  const agent = await actions.createAgent({
    commission:2.5, 
    products: [
      {standardPrice:5,
        items: [
          {quantity:2},
          {quantity:5},
        ]
      }
    ]});
    const products = await models.product.findMany({ where: { agentId: agent.id }});

    const items = await models.item.findMany({ where: { productId: products[0].id }, orderBy: { total: "asc"}});

  expect(items[0].total).toEqual(12.5);
  expect(items[1].total).toEqual(27.5);
  expect(items).length(2);

});
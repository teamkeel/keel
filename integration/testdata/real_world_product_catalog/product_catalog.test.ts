import { actions, models } from "@teamkeel/testing";
import { test, expect, beforeAll } from "vitest";

let appleIdentity;
let samsungIdentity;

beforeAll(async () => {
  appleIdentity = await models.identity.create({ email: "apple@myshop.co.uk" });
  samsungIdentity = await models.identity.create({
    email: "samsung@myshop.co.uk",
  });

  const apple = await actions.addBrand({
    name: "Apple",
    code: "APL",
    products: [
      {
        name: "Apple Macbook Pro",
        barcode: "346123498729",
        productCode: "00001",
      },
      {
        name: "Apple iPhone 15",
        barcode: "57227788834",
        productCode: "IPH15",
      },
      {
        name: "Apple Watch",
        barcode: "0366026378333",
        productCode: "00002",
      },
      {
        name: "Apple iPhone 12",
        barcode: "21109374622",
        productCode: "IPH12",
      },
    ],
  });

  await models.user.create({ identityId: appleIdentity.id, brandId: apple.id });

  const appleMacBook = await actions.getProduct({
    brandCode: "APL",
    productCode: "00001",
  });

  const appleWatch = await actions.getProduct({
    brandCode: "APL",
    productCode: "00002",
  });

  const appleIPhone = await actions.getProduct({
    brandCode: "APL",
    productCode: "IPH12",
  });

  const samsung = await actions.addBrand({
    name: "Samsung",
    code: "SMSG",
    products: [
      {
        name: "Samsung Air Pods",
        barcode: "7943323452",
        productCode: "TS088",
      },
      {
        name: "Samsung Galaxy S23",
        barcode: "12260000654",
        productCode: "00001",
      },
      {
        name: "Samsung 55inch Neo QLED 4K",
        barcode: "82565464456",
        productCode: "00002",
      },
    ],
  });

  await models.user.create({
    identityId: samsungIdentity.id,
    brandId: samsung.id,
  });

  const samsungGalaxy = await actions.getProduct({
    brandCode: "SMSG",
    productCode: "00001",
  });

  const samsungNeoTv = await actions.getProduct({
    brandCode: "SMSG",
    productCode: "00002",
  });

  const abcElectronics = await actions.addSupplier({
    name: "ABC Electronics",
    code: "ABC",
    catalog: [
      {
        product: { id: appleMacBook!.id },
        supplierSku: "app001",
        price: 2499,
        stockCount: 1,
      },
      {
        product: { id: appleWatch!.id },
        supplierSku: "app002",
        price: 999,
        stockCount: 35,
      },
      {
        product: { id: samsungGalaxy!.id },
        supplierSku: "sam001",
        price: 199,
        stockCount: 56,
      },
      {
        product: { id: samsungNeoTv!.id },
        supplierSku: "sam003",
        price: 759,
        stockCount: 3,
      },
      {
        product: { id: appleIPhone!.id },
        supplierSku: "iphone",
        price: 1049,
        stockCount: 1,
      },
    ],
  });

  const phoneShop = await actions.addSupplier({
    name: "Phones Galore",
    code: "GALORE",
    catalog: [
      {
        product: { id: samsungGalaxy!.id },
        supplierSku: "galaxy",
        price: 209,
        stockCount: 10,
      },
      {
        product: { id: appleMacBook!.id },
        supplierSku: "app001",
        price: 1999,
        stockCount: 20,
      },
      {
        product: { id: appleIPhone!.id },
        supplierSku: "iphone",
        price: 1299,
        stockCount: 5,
      },
    ],
  });
});

test("create product with duplicated product code - invalid inputs response", async () => {
  const samsung = await actions.getBrand({ code: "SMSG" });
  expect(samsung).not.toBeNull();

  await expect(
    actions.withIdentity(samsungIdentity).createProduct({
      name: "Pad",
      barcode: "3823323598",
      productCode: "TS088",
      brand: {
        id: samsung!.id,
      },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message:
      "the values for the unique composite fields (brandId, productCode) must be unique",
  });

  const product = await actions.withIdentity(samsungIdentity).createProduct({
    name: "Pad",
    barcode: "3823323598",
    productCode: "ZZ000",
    brand: {
      id: samsung!.id,
    },
  });
  expect(product).not.toBeNull();
});

test("create product without brand permissions - permission error", async () => {
  const samsung = await actions.getBrand({ code: "SMSG" });
  expect(samsung).not.toBeNull();

  await expect(
    actions.withIdentity(appleIdentity).createProduct({
      name: "Pad",
      barcode: "398328223",
      productCode: "KA222",
      brand: {
        id: samsung!.id,
      },
    })
  ).toHaveError({
    code: "ERR_PERMISSION_DENIED",
    message: "not authorized to access",
  });
});

test("create product with brand permissions - product created", async () => {
  const samsung = await actions.getBrand({ code: "SMSG" });
  expect(samsung).not.toBeNull();

  const product = await actions.withIdentity(samsungIdentity).createProduct({
    name: "Samsung Watch",
    barcode: "29387928",
    productCode: "SMW111",
    brand: {
      id: samsung!.id,
    },
  });

  expect(product).not.toBeNull();
  expect(product.barcode).toEqual("29387928");
});

test("create product with duplicate barcode - invalid inputs response", async () => {
  const samsung = await actions.getBrand({ code: "SMSG" });
  expect(samsung).not.toBeNull();

  await expect(
    actions.withIdentity(samsungIdentity).createProduct({
      name: "Pad",
      barcode: "3823323598",
      productCode: "LSI22",
      brand: {
        id: samsung!.id,
      },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the value for the unique field 'barcode' must be unique",
  });
});

test("get product by composite unique - product returned", async () => {
  const brand = await actions.getBrand({
    code: "APL",
  });

  expect(brand).not.toBeNull();
  expect(brand?.name).toEqual("Apple");

  const macbook = await actions.getProduct({
    productCode: "00001",
    brandCode: brand!.code,
  });

  expect(macbook).not.toBeNull();
  expect(macbook?.name).toEqual("Apple Macbook Pro");
});

test("get product by composite unique - product not found", async () => {
  const brand = await actions.getBrand({
    code: "APL",
  });

  expect(brand).not.toBeNull();
  expect(brand?.name).toEqual("Apple");

  const macbook = await actions.getProduct({
    productCode: "notfound",
    brandCode: brand!.code,
  });

  expect(macbook).toBeNull();
});

test("get catalog item by composite unique down a relationship - item returned", async () => {
  const brand = await actions.getBrand({
    code: "APL",
  });

  expect(brand).not.toBeNull();
  expect(brand?.name).toEqual("Apple");

  const macbook = await actions.getProduct({
    productCode: "00001",
    brandCode: brand!.code,
  });

  expect(macbook).not.toBeNull();
  expect(macbook?.name).toEqual("Apple Macbook Pro");

  const macbookFromAbc = await actions.getCatalogItem({
    supplierCode: "ABC",
    productProductCode: macbook!.productCode,
    productBrandCode: brand!.code,
  });

  expect(macbookFromAbc).not.toBeNull();
  expect(macbookFromAbc?.price).toEqual(2499);

  const macbookFromGalore = await actions.getCatalogItem({
    supplierCode: "GALORE",
    productProductCode: macbook!.productCode,
    productBrandCode: brand!.code,
  });

  expect(macbookFromGalore).not.toBeNull();
  expect(macbookFromGalore?.price).toEqual(1999);
});

test("get catalog item by composite unique down a relationship - item not found", async () => {
  const brand = await actions.getBrand({
    code: "APL",
  });

  expect(brand).not.toBeNull();
  expect(brand?.name).toEqual("Apple");

  const macbook = await actions.getProduct({
    productCode: "00001",
    brandCode: brand!.code,
  });

  expect(macbook).not.toBeNull();
  expect(macbook?.name).toEqual("Apple Macbook Pro");

  const macbookFromAbc = await actions.getCatalogItem({
    supplierCode: "ABC",
    productProductCode: macbook!.productCode,
    productBrandCode: "notfound",
  });

  expect(macbookFromAbc).toBeNull();
});

test("get brand for a product - brand returned", async () => {
  const samsung = await actions.getBrandForProduct({
    productsBarcode: "7943323452",
  });

  expect(samsung).not.toBeNull();
  expect(samsung?.name).toEqual("Samsung");
});

test("get brand for a product - brand not found", async () => {
  const samsung = await actions.getBrandForProduct({
    productsBarcode: "notfound",
  });

  expect(samsung).toBeNull();
});

test("list suppliers with a certain product - suppliers returned", async () => {
  const suppliers = await actions.suppliersWithProduct({
    where: {
      catalog: {
        product: {
          barcode: {
            equals: "346123498729",
          },
        },
      },
    },
  });

  expect(suppliers.pageInfo.count).toEqual(2);
  expect(suppliers.results[0].name).toEqual("ABC Electronics");
  expect(suppliers.results[1].name).toEqual("Phones Galore");
});

test("deactive product without brand access - permission error", async () => {
  await expect(
    actions.withIdentity(appleIdentity).deactivateProduct({
      where: {
        brandCode: "SMSG",
        productCode: "00002",
      },
    })
  ).toHaveAuthorizationError();
});

test("deactive product - product deactived", async () => {
  const product = await actions
    .withIdentity(samsungIdentity)
    .deactivateProduct({
      where: {
        brandCode: "SMSG",
        productCode: "00002",
      },
    });

  expect(product).not.toBeNull();
  expect(product.isActive).toBeFalsy();

  const getProduct = await actions.getProduct({
    brandCode: "SMSG",
    productCode: "00002",
  });

  expect(getProduct).toBeNull();

  const suppliers = await actions.suppliersWithProduct({
    where: {
      catalog: {
        product: {
          barcode: {
            equals: product.barcode,
          },
        },
      },
    },
  });

  expect(suppliers.pageInfo.count).toEqual(0);
});

test("list suppliers with a certain product - product doesnt exist", async () => {
  const suppliers = await actions.suppliersWithProduct({
    where: {
      catalog: {
        product: {
          barcode: {
            equals: "notexists",
          },
        },
      },
    },
  });

  expect(suppliers.pageInfo.count).toEqual(0);
});

test("search catalog - catalog items returned", async () => {
  const appleItems = await actions.searchCatalog({
    where: {
      product: {
        name: {
          contains: "Apple",
        },
        productCode: {
          contains: "IPH",
        },
      },
    },
    orderBy: [{ price: "asc" }, { stockCount: "desc" }],
  });

  expect(appleItems.pageInfo.count).toEqual(2);
  expect(appleItems.results[0].supplierSku).toEqual("iphone");
  expect(appleItems.results[0].price).toEqual(1049);

  expect(appleItems.results[1].supplierSku).toEqual("iphone");
  expect(appleItems.results[1].price).toEqual(1299);
});

test("delete product without brand access - permission error", async () => {
  await expect(
    actions.withIdentity(samsungIdentity).deleteProduct({
      brandCode: "APL",
      productCode: "IPH15",
    })
  ).toHaveAuthorizationError();
});

test("delete product by composite uniques - deleted", async () => {
  const product = await actions.getProduct({
    brandCode: "APL",
    productCode: "IPH15",
  });

  expect(product).not.toBeNull();

  const productId = await actions.withIdentity(appleIdentity).deleteProduct({
    brandCode: "APL",
    productCode: "IPH15",
  });

  expect(productId).toEqual(product?.id);

  const appleItems = await actions.searchCatalog({
    where: {
      product: {
        name: {
          contains: "Apple",
        },
        productCode: {
          equals: "IPH15",
        },
      },
    },
    orderBy: [{ price: "asc" }, { stockCount: "desc" }],
  });

  expect(appleItems.pageInfo.count).toEqual(0);
});

test("delete product of inactive brand - not found", async () => {
  const product = await actions.getProduct({
    brandCode: "SMSG",
    productCode: "TS088",
  });

  expect(product).not.toBeNull();

  await models.brand.update({ code: "SMSG" }, { isActive: false });

  await expect(
    actions.withIdentity(samsungIdentity).deleteProduct({
      brandCode: "SMSG",
      productCode: "TS088",
    })
  ).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

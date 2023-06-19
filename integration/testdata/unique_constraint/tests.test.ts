import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("create operation - unique constraint error", async () => {
    await expect(
        actions.createProduct({ name: "Mountain Bike", sku: "MB001", supplierCode: "S1" })
    ).not.toHaveError({});

    await expect(
        actions.createProduct({ name: "Mountain Bike", sku: "MB001", supplierCode: "S2" })
    ).not.toHaveError({});

    await expect(
        actions.createProduct({ name: "Mountain Bike", sku: "MB001", supplierCode: "S1" })
    ).toHaveError({
        code: "ERR_INVALID_INPUT",
        message: "unique composite (name, supplierCode) can only contain unique values",
    });
});

test("update operation - unique constraint error", async () => {
    await models.product.create({ name: "Mountain Bike", sku: "MB001", supplierCode: "S1" });
    const {id } = await models.product.create({ name: "Mountain Bike", sku: "MB001", supplierCode: "S2" });

    await expect(
        actions.updateProduct({ where: { id: id }, values: { sku: "MB001", supplierCode: "S1" }})
    ).toHaveError({
        code: "ERR_INVALID_INPUT",
        message: "unique composite (name, supplierCode) can only contain unique values",
    });

    await expect(
        actions.updateProduct({ where: { id: id }, values: { supplierCode: "S1" }})
    ).toHaveError({
        code: "ERR_INVALID_INPUT",
        message: "unique composite (name, supplierCode) can only contain unique values",
    });
});

test("create function - unique constraint error", async () => {
    await expect(
        actions.createProductFn({ name: "Mountain Bike", sku: "MB001", supplierCode: "S1" })
    ).not.toHaveError({});

    await expect(
        actions.createProductFn({ name: "Mountain Bike", sku: "MB001", supplierCode: "S2" })
    ).not.toHaveError({});

    await expect(
        actions.createProductFn({ name: "Mountain Bike", sku: "MB001", supplierCode: "S1" })
    ).toHaveError({
        code: "ERR_INVALID_INPUT",
        message: "unique composite (name, supplierCode) can only contain unique values",
    });
});

test("update function - unique constraint error", async () => {
    await models.product.create({ name: "Mountain Bike", sku: "MB001", supplierCode: "S1" });
    const {id } = await models.product.create({ name: "Mountain Bike", sku: "MB001", supplierCode: "S2" });

    await expect(
        actions.updateProductFn({ where: { id: id }, values: { sku: "MB001", supplierCode: "S1" }})
    ).toHaveError({
        code: "ERR_INVALID_INPUT",
        message: "unique composite (name, supplierCode) can only contain unique values",
    });

    await expect(
        actions.updateProductFn({ where: { id: id }, values: { supplierCode: "S1" }})
    ).toHaveError({
        code: "ERR_INVALID_INPUT",
        message: "unique composite (name, supplierCode) can only contain unique values",
    });
});
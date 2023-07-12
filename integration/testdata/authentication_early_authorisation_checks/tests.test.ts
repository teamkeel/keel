import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("create op - ctx.isAuthenticated", async () => {
  await expect(
    actions.createBook({ title: "My Book" })
  ).toHaveAuthorizationError();
});

test("create op - ctx.isAuthenticated == false", async () => {
  await expect(actions.createBook2({ title: "My Book" })).toHaveError({
    message: "field 'lastUpdatedById' cannot be null",
  });
});

test("create op - database check", async () => {
  await expect(actions.createBook3({ title: "My Book" })).toHaveError({
    message: "field 'lastUpdatedById' cannot be null",
  });
});

test("update op - ctx.isAuthenticated", async () => {
  await expect(
    actions.updateBook({ where: { id: "123" }, values: { title: "My Book" } })
  ).toHaveAuthorizationError();
});

test("update op - ctx.isAuthenticated == false", async () => {
  await expect(
    actions.updateBook2({ where: { id: "123" }, values: { title: "My Book" } })
  ).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("update op - database check", async () => {
  await expect(
    actions.updateBook3({ where: { id: "123" }, values: { title: "My Book" } })
  ).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("get op - ctx.isAuthenticated", async () => {
  await expect(actions.getBook({ id: "123" })).toHaveAuthorizationError();
});

test("get op - ctx.isAuthenticated == false", async () => {
  const book = await actions.getBook2({ id: "123" });
  expect(book).toBeNull();
});

test("get op - database check", async () => {
  const book = await actions.getBook3({ id: "123" });
  expect(book).toBeNull();
});

test("delete op - ctx.isAuthenticated", async () => {
  await expect(actions.deleteBook({ id: "123" })).toHaveAuthorizationError();
});

test("delete op - ctx.isAuthenticated == false", async () => {
  await expect(actions.deleteBook2({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete op - database check", async () => {
  await expect(actions.deleteBook3({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("list op - ctx.isAuthenticated", async () => {
  await expect(actions.listBook()).toHaveAuthorizationError();
});

test("list op - ctx.isAuthenticated == false", async () => {
  const books = await actions.listBook2();
  expect(books.results).toHaveLength(0);
});

test("list op - database check", async () => {
  const books = await actions.listBook3();
  expect(books.results).toHaveLength(0);
});

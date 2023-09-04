import { test, expect, beforeEach } from "vitest";
import { actions, models, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("create action - permitted, @set to null - ERR_INVALID_INPUT", async () => {
  await expect(actions.createPermitted({ title: "My Book" })).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "field 'lastUpdatedById' cannot be null",
  });
});

test("create action - permitted, lookup fail - ERR_INVALID_INPUT", async () => {
  await expect(
    actions.createPermittedNoSet({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'lastUpdatedById' does not exist",
  });
});

test("create action - not permitted, @set identity - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createNotPermitted({ title: "My Book" })
  ).toHaveAuthorizationError();

  await expect(await models.book.findMany()).toHaveLength(0);
});

test("create action - not authenticated, @set identity - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createIsAuthenticated({ title: "My Book" })
  ).toHaveAuthorizationError();

  await expect(await models.book.findMany()).toHaveLength(0);
});

test("create action - database permission, @set to null - ERR_INVALID_INPUT", async () => {
  await expect(actions.createDbPermission({ title: "My Book" })).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "field 'lastUpdatedById' cannot be null",
  });

  await expect(await models.book.findMany()).toHaveLength(0);
});

test("create action - database permission, lookup failed - ERR_INVALID_INPUT", async () => {
  await expect(
    actions.createDbPermissionNoSet({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'lastUpdatedById' does not exist",
  });

  await expect(await models.book.findMany()).toHaveLength(0);
});

test("create action - database permission, no identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });

  await expect(
    actions.createDbPermissionNoSet({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveAuthorizationError();

  await expect(await models.book.findMany()).toHaveLength(0);
});

test("create action - database permission, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions.withIdentity(wrongIdentity).createDbPermissionNoSet({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveAuthorizationError();

  await expect(await models.book.findMany()).toHaveLength(0);
});

test("create function - permitted, lookup fail - ERR_INVALID_INPUT", async () => {
  await expect(
    actions.createPermittedFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'lastUpdatedById' does not exist",
  });
});

test("create function - not permitted, null identity - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createNotPermittedFn({ title: "My Book" })
  ).toHaveAuthorizationError();
});

test("create function - not permitted, lookup fail - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createNotPermittedFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveAuthorizationError();
});

test("create function - not authenticated, null identity - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createIsAuthenticatedFn({ title: "My Book" })
  ).toHaveAuthorizationError();
});

test("create function - not authenticated, lookup fail - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createIsAuthenticatedFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveAuthorizationError();
});

test("create function - database permission, lookup fail - ERR_INVALID_INPUT", async () => {
  await expect(
    actions.createDbPermissionFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'lastUpdatedById' does not exist",
  });
});

test("create function - database permission, no identity - ERR_INVALID_INPUT", async () => {
  await models.identity.create({ id: "someId" });

  await expect(
    actions.createDbPermissionFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'lastUpdatedById' does not exist",
  });
});

test("create function - database permission, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions.withIdentity(wrongIdentity).createDbPermissionFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveAuthorizationError();

  await expect(await models.book.findMany()).toHaveLength(0);
});

test("update action - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.updateNotPermitted({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("update action - not permitted, id exists - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.updateNotPermitted({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("update action - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(
    actions.updatePermitted({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("update action - not permitted, lookup fail - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.updateNotPermitted({
      where: { id: "123" },
      values: { title: "My Book", lastUpdatedBy: { id: "no match" } },
    })
  ).toHaveAuthorizationError();
});

test("update action - permitted, lookup fail - ERR_INVALID_INPUT", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.updatePermitted({
      where: { id: "123" },
      values: { title: "My Book", lastUpdatedBy: { id: "no match" } },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'lastUpdatedById' does not exist",
  });
});

test("update action - database check, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(
    actions.updateDbPermission({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("update action - database check, no identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.updateDbPermission({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("update action - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });
  await expect(
    actions.withIdentity(wrongIdentity).updateDbPermission({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("update function - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.updateNotPermittedFn({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("update function - not permitted, id exists - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.updateNotPermittedFn({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("update function - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(
    actions.updatePermittedFn({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("update function - not permitted, lookup fail - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.updateNotPermittedFn({
      where: { id: "123" },
      values: { title: "My Book", lastUpdatedBy: { id: "no match" } },
    })
  ).toHaveAuthorizationError();
});

test("update function - permitted, lookup fail - ERR_INVALID_INPUT", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.updatePermittedFn({
      where: { id: "123" },
      values: { title: "My Book", lastUpdatedBy: { id: "no match" } },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the record referenced in field 'lastUpdatedById' does not exist",
  });
});

test("update function - database check, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(
    actions.updateDbPermissionFn({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("update function - database check, no identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.updateDbPermissionFn({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("update function - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });
  await expect(
    actions.withIdentity(wrongIdentity).updateDbPermissionFn({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("get action - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.getNotPermitted({ id: "123" })
  ).toHaveAuthorizationError();
});

test("get action - permitted, id not exists - null returned", async () => {
  const book = await actions.getPermitted({ id: "123" });
  expect(book).toBeNull();
});

test("get action - database check, id not exists - null returned", async () => {
  const book = await actions.getDbPermission({ id: "123" });
  expect(book).toBeNull();
});

test("get action - database check, no identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.getDbPermission({ id: "123" })
  ).toHaveAuthorizationError();
});

test("get action - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions.withIdentity(wrongIdentity).getDbPermission({ id: "123" })
  ).toHaveAuthorizationError();
});

test("get function - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.getNotPermittedFn({ id: "123" })
  ).toHaveAuthorizationError();
});

test("get function - permitted, id not exists - null returned", async () => {
  const book = await actions.getPermittedFn({ id: "123" });
  expect(book).toBeNull();
});

test("get function - database check, id not exists - null returned", async () => {
  const book = await actions.getDbPermissionFn({ id: "123" });
  expect(book).toBeNull();
});

test("get function - database check, no identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.getDbPermissionFn({ id: "123" })
  ).toHaveAuthorizationError();
});

test("get function - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions.withIdentity(wrongIdentity).getDbPermissionFn({ id: "123" })
  ).toHaveAuthorizationError();
});

test("delete action - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.deleteNotPermitted({ id: "123" })
  ).toHaveAuthorizationError();
});

test("delete action - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(actions.deletePermitted({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete action - database check, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(actions.deleteDbPermission({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete action - database check, no identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.deleteDbPermission({ id: "123" })
  ).toHaveAuthorizationError();
});

test("delete action - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions.withIdentity(wrongIdentity).deleteDbPermission({ id: "123" })
  ).toHaveAuthorizationError();
});

test("delete function - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.deleteNotPermittedFn({ id: "123" })
  ).toHaveAuthorizationError();
});

test("delete function - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(actions.deletePermittedFn({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete function - database check, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(actions.deleteDbPermissionFn({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete function - database check, no identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.deleteDbPermissionFn({ id: "123" })
  ).toHaveAuthorizationError();
});

test("delete function - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions.withIdentity(wrongIdentity).deleteDbPermissionFn({ id: "123" })
  ).toHaveAuthorizationError();
});

test("list action - not permitted - ERR_PERMISSION_DENIED", async () => {
  await expect(actions.listNotPermitted()).toHaveAuthorizationError();
});

test("list action - permitted, no rows - empty result", async () => {
  const books = await actions.listPermitted();
  expect(books.results).toHaveLength(0);
});

test("list action - database check, with identity, no rows - empty result", async () => {
  const identity = await models.identity.create({
    id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  const books = await actions.withIdentity(identity).listDbPermission();
  expect(books.results).toHaveLength(0);
});

test("list action - database check, no identity, no rows - empty result", async () => {
  const books = await actions.listDbPermission();
  expect(books.results).toHaveLength(0);
});

test("list action - database check, no identity, with rows - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(actions.listDbPermission()).toHaveAuthorizationError();
});

test("list action - database check, wrong identity, with rows - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions.withIdentity(wrongIdentity).listDbPermission()
  ).toHaveAuthorizationError();
});

test("list action - database check, correct identity, with rows - rows returned", async () => {
  const identity = await models.identity.create({
    id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.withIdentity(identity).listDbPermission()
  ).not.toHaveAuthorizationError();
});

test("list function - not permitted - ERR_PERMISSION_DENIED", async () => {
  await expect(actions.listNotPermittedFn()).toHaveAuthorizationError();
});

test("list function - permitted, no rows - empty result", async () => {
  const books = await actions.listPermittedFn();
  expect(books.results).toHaveLength(0);
});

test("list function - database check, with identity, no rows - empty result", async () => {
  const identity = await models.identity.create({
    id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  const books = await actions.withIdentity(identity).listDbPermissionFn();
  expect(books.results).toHaveLength(0);
});

test("list function - database check, no identity, no rows - empty result", async () => {
  const books = await actions.listDbPermissionFn();
  expect(books.results).toHaveLength(0);
});

test("list function - database check, no identity, with rows - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(actions.listDbPermissionFn()).toHaveAuthorizationError();
});

test("list function - database check, wrong identity, with rows - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions.withIdentity(wrongIdentity).listDbPermissionFn()
  ).toHaveAuthorizationError();
});

test("list function - database check, correct identity, with rows - rows returned", async () => {
  const identity = await models.identity.create({
    id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });
  await models.book.create({
    id: "123",
    title: "Harry Potter",
    lastUpdatedById: "2PvOAtybZaxSzf1WGNKaWd5BZ0R",
  });

  await expect(
    actions.withIdentity(identity).listDbPermissionFn()
  ).not.toHaveAuthorizationError();
});

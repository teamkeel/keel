import { test, expect, beforeEach } from "vitest";
import { actions, models, resetDatabase } from "@teamkeel/testing";
import { s } from "vitest/dist/index-50755efe";

beforeEach(resetDatabase);

// CREATE OPERATIONS

test("create op - permitted, @set to null - ERR_INVALID_INPUT", async () => {
  await expect(actions.createPermitted({ title: "My Book" })).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "field 'lastUpdatedById' cannot be null",
  });
});

test("create op - permitted, lookup fail - ERR_INVALID_INPUT", async () => {
  await expect(
    actions.createPermittedNoSet({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message:
      "the relationship lookup for field 'lastUpdatedById' does not exist",
  });
});

test("create op - not permitted, @set identity - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createNotPermitted({ title: "My Book" })
  ).toHaveAuthorizationError();
});

test("create op - not authenticated, @set identity - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createIsAuthenticated({ title: "My Book" })
  ).toHaveAuthorizationError();
});

test("create op - database permission, @set to null - ERR_INVALID_INPUT", async () => {
  await expect(actions.createDbPermission({ title: "My Book" })).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "field 'lastUpdatedById' cannot be null",
  });
});

test("create op - database permission, lookup failed - ERR_INVALID_INPUT", async () => {
  await expect(
    actions.createDbPermissionNoSet({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message:
      "the relationship lookup for field 'lastUpdatedById' does not exist",
  });
});

test("create op - database permission, no identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });

  await expect(
    actions.createDbPermissionNoSet({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveAuthorizationError();
});

test("create op - database permission, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions
      .withIdentity(wrongIdentity)
      .createDbPermissionNoSet({
        title: "My Book",
        lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
      })
  ).toHaveAuthorizationError();
});

// CREATE FUNCTIONS

test("create func - permitted, lookup fail - ERR_INVALID_INPUT", async () => {
  await expect(
    actions.createPermittedFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message:
      "the relationship lookup for field 'lastUpdatedById' does not exist",
  });
});

test("create func - not permitted, null identity - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createNotPermittedFn({ title: "My Book" })
  ).toHaveAuthorizationError();
});

test("create func - not permitted, lookup fail - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createNotPermittedFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveAuthorizationError();
});

test("create func - not authenticated, null identity - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createIsAuthenticatedFn({ title: "My Book" })
  ).toHaveAuthorizationError();
});

test("create func - not authenticated, lookup fail - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.createIsAuthenticatedFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveAuthorizationError();
});

test("create func - database permission, lookup fail - ERR_INVALID_INPUT", async () => {
  await expect(
    actions.createDbPermissionFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message:
      "the relationship lookup for field 'lastUpdatedById' does not exist",
  });
});

test("create func - database permission, no identity - ERR_INVALID_INPUT", async () => {
  await models.identity.create({ id: "someId" });

  await expect(
    actions.createDbPermissionFn({
      title: "My Book",
      lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message:
      "the relationship lookup for field 'lastUpdatedById' does not exist",
  });
});

test("create func - database permission, wrong identity - ERR_PERMISSION_DENIED", async () => {
  await models.identity.create({ id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" });
  const wrongIdentity = await models.identity.create({
    id: "2Qb2ItMXLmNXDun8tk1z75mbZhj",
  });

  await expect(
    actions
      .withIdentity(wrongIdentity)
      .createDbPermissionFn({
        title: "My Book",
        lastUpdatedBy: { id: "2PvOAtybZaxSzf1WGNKaWd5BZ0R" },
      })
  ).toHaveAuthorizationError();
});

// UPDATE OPERATIONS

test("update op - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.updateNotPermitted({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("update op - not permitted, id exists - ERR_PERMISSION_DENIED", async () => {
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

test("update op - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
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

test("update op - not permitted, lookup fail - ERR_PERMISSION_DENIED", async () => {
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

test("update op - permitted, lookup fail - ERR_INVALID_INPUT", async () => {
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
    message:
      "the relationship lookup for field 'lastUpdatedById' does not exist",
  });
});

test("update op - database check, id not exists - ERR_RECORD_NOT_FOUND", async () => {
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

test("update op - database check, no identity - ERR_PERMISSION_DENIED", async () => {
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

test("update op - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
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
    actions
      .withIdentity(wrongIdentity)
      .updateDbPermission({
        where: { id: "123" },
        values: { title: "My Book" },
      })
  ).toHaveAuthorizationError();
});

// UPDATE FUNCTIONS

test("update func - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.updateNotPermittedFn({
      where: { id: "123" },
      values: { title: "My Book" },
    })
  ).toHaveAuthorizationError();
});

test("update func - not permitted, id exists - ERR_PERMISSION_DENIED", async () => {
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

test("update func - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
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

test("update func - not permitted, lookup fail - ERR_PERMISSION_DENIED", async () => {
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

test("update func - permitted, lookup fail - ERR_INVALID_INPUT", async () => {
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
    message:
      "the relationship lookup for field 'lastUpdatedById' does not exist",
  });
});

test("update func - database check, id not exists - ERR_RECORD_NOT_FOUND", async () => {
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

test("update func - database check, no identity - ERR_PERMISSION_DENIED", async () => {
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

test("update func - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
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
    actions
      .withIdentity(wrongIdentity)
      .updateDbPermissionFn({
        where: { id: "123" },
        values: { title: "My Book" },
      })
  ).toHaveAuthorizationError();
});

// GET OPERATIONS

test("get op - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.getNotPermitted({ id: "123" })
  ).toHaveAuthorizationError();
});

test("get op - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  const book = await actions.getPermitted({ id: "123" });
  expect(book).toBeNull();
});

test("get op - permitted, id not exists - null returned", async () => {
  const book = await actions.getPermitted({ id: "123" });
  expect(book).toBeNull();
});

test("get op - database check, id not exists - null returned", async () => {
  const book = await actions.getDbPermission({ id: "123" });
  expect(book).toBeNull();
});

test("get op - database check, no identity - ERR_PERMISSION_DENIED", async () => {
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

test("get op - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
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

// GET FUNCTIONS

test("get func - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.getNotPermittedFn({ id: "123" })
  ).toHaveAuthorizationError();
});

test("get func - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  const book = await actions.getPermittedFn({ id: "123" });
  expect(book).toBeNull();
});

test("get func - permitted, id not exists - null returned", async () => {
  const book = await actions.getPermittedFn({ id: "123" });
  expect(book).toBeNull();
});

test("get func - database check, id not exists - null returned", async () => {
  const book = await actions.getDbPermissionFn({ id: "123" });
  expect(book).toBeNull();
});

test("get func - database check, no identity - ERR_PERMISSION_DENIED", async () => {
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

test("get func - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
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

// DELETE OPERATIONS

test("delete op - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.deleteNotPermitted({ id: "123" })
  ).toHaveAuthorizationError();
});

test("delete op - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(actions.deletePermitted({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete op - database check, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(actions.deleteDbPermission({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete op - database check, no identity - ERR_PERMISSION_DENIED", async () => {
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

test("delete op - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
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

// DELETE FUNCTIONS

test("delete func - not permitted, id not exists - ERR_PERMISSION_DENIED", async () => {
  await expect(
    actions.deleteNotPermittedFn({ id: "123" })
  ).toHaveAuthorizationError();
});

test("delete func - permitted, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(actions.deletePermittedFn({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete func - database check, id not exists - ERR_RECORD_NOT_FOUND", async () => {
  await expect(actions.deleteDbPermissionFn({ id: "123" })).toHaveError({
    code: "ERR_RECORD_NOT_FOUND",
    message: "record not found",
  });
});

test("delete func - database check, no identity - ERR_PERMISSION_DENIED", async () => {
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

test("delete func - database check, wrong identity - ERR_PERMISSION_DENIED", async () => {
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

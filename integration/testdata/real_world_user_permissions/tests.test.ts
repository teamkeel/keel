import { actions, models } from "@teamkeel/testing";
import { Role, Permission, User, Identity } from "@teamkeel/sdk";
import { test, expect, beforeAll } from "vitest";

let roleSupervisor: Role;
let roleClerk: Role;
let roleAdmin: Role;
let permAccountCreate: Permission;
let permAccountRead: Permission;
let permAccountGet: Permission;
let permAccountDelete: Permission;
let supervisorIdentity: Identity;
let clerkIdentity: Identity;
let supervisor: User;
let clerk: User;

beforeAll(async () => {
  roleAdmin = await models.role.create({
    name: "admin",
  });

  roleSupervisor = await models.role.create({
    name: "supervisor",
  });

  roleClerk = await models.role.create({
    name: "clerk",
  });

  permAccountDelete = await models.permission.create({
    name: "account:delete",
  });

  permAccountCreate = await models.permission.create({
    name: "account:create",
  });

  permAccountRead = await models.permission.create({
    name: "account:list",
  });

  permAccountGet = await models.permission.create({
    name: "account:get",
  });

  await models.rolePermission.create({
    roleId: roleSupervisor.id,
    permissionId: permAccountCreate.id,
  });

  await models.rolePermission.create({
    roleId: roleSupervisor.id,
    permissionId: permAccountRead.id,
  });

  await models.rolePermission.create({
    roleId: roleSupervisor.id,
    permissionId: permAccountGet.id,
  });

  await models.rolePermission.create({
    roleId: roleClerk.id,
    permissionId: permAccountRead.id,
  });

  await models.rolePermission.create({
    roleId: roleClerk.id,
    permissionId: permAccountRead.id,
  });

  await models.rolePermission.create({
    roleId: roleAdmin.id,
    permissionId: permAccountDelete.id,
  });

  supervisorIdentity = await models.identity.create({
    email: "supervisor@test.com",
    password: "foo",
  });
  supervisor = await models.user.create({
    identityId: supervisorIdentity.id,
  });

  clerkIdentity = await models.identity.create({
    email: "clerk@test.com",
    password: "foo",
  });
  clerk = await models.user.create({
    identityId: clerkIdentity.id,
  });

  const clerkIdentity2 = await models.identity.create({
    email: "clerk@test.com",
    password: "foo",
  });
  await models.user.create({
    identityId: clerkIdentity2.id,
  });

  await models.userRole.create({
    userId: supervisor.id,
    roleId: roleSupervisor.id,
  });

  await models.userRole.create({
    userId: supervisor.id,
    roleId: roleAdmin.id,
  });

  await models.userRole.create({
    userId: clerk.id,
    roleId: roleClerk.id,
  });
});

test("create account without user - not authorised", async () => {
  await expect(
    actions.createAccount({
      name: "test",
    })
  ).toHaveAuthorizationError();
});

test("create account by clerk - not authorised", async () => {
  await expect(
    actions.withIdentity(clerkIdentity).createAccount({
      name: "test",
    })
  ).toHaveAuthorizationError();
});

test("create account by supervisor - authorised", async () => {
  const account = await actions.withIdentity(supervisorIdentity).createAccount({
    name: "test",
  });

  expect(account).toBeDefined();
  expect(account.name).toBe("test");
});

test("list accounts without user - not authorised", async () => {
  await expect(actions.listAccount()).toHaveAuthorizationError();
});

test("list accounts by clerk - authorised", async () => {
  const accounts = await actions.withIdentity(clerkIdentity).listAccount();

  expect(accounts).toBeDefined();
});

test("list accounts by supervisor - authorised", async () => {
  const accounts = await actions.withIdentity(supervisorIdentity).listAccount();

  expect(accounts).toBeDefined();
});

test("create account function without user - not authorised", async () => {
  await expect(
    actions.createAccountFn({
      name: "test",
    })
  ).toHaveAuthorizationError();
});

test("create account function by clerk - not authorised", async () => {
  await expect(
    actions.withIdentity(clerkIdentity).createAccountFn({
      name: "test",
    })
  ).toHaveAuthorizationError();
});

test("create account function by supervisor - authorised", async () => {
  const account = await actions
    .withIdentity(supervisorIdentity)
    .createAccountFn({
      name: "test",
    });

  expect(account).toBeDefined();
  expect(account.name).toBe("test");
});

test("list accounts function without user - not authorised", async () => {
  await expect(actions.listAccountFn()).toHaveAuthorizationError();
});

test("list accounts function by clerk - authorised", async () => {
  const accounts = await actions.withIdentity(clerkIdentity).listAccountFn();

  expect(accounts.results).toBeDefined();
});

test("list accounts function by supervisor - authorised", async () => {
  const accounts = await actions
    .withIdentity(supervisorIdentity)
    .listAccountFn();

  expect(accounts.results).toBeDefined();
});

import { actions, models } from "@teamkeel/testing";
import { Role, Permission, User, Identity } from "@teamkeel/sdk"
import { test, expect, beforeAll } from "vitest";

let roleSupervisor: Role;
let roleClerk: Role;
let permAccountCreate: Permission;
let permAccountRead: Permission;
let supervisorIdentity: Identity;
let clerkIdentity: Identity;
let supervisor: User;
let clerk: User;


beforeAll(async () => {
  roleSupervisor = await models.role.create({
    name: "supervisor",
  });

  roleClerk = await models.role.create({
    name: "clerk",
  });

  permAccountCreate = await models.permission.create({
    name: "account:create",
  });

  permAccountRead = await models.permission.create({
    name: "account:read",
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
    roleId: roleClerk.id,
    permissionId: permAccountRead.id,
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

  await models.userRole.create({
    userId: supervisor.id,
    roleId: roleSupervisor.id,
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
  await expect(
    actions.createAccount({
      name: "test",
    })
  ).toHaveAuthorizationError();

});


test("list accounts by clerk - authorised", async () => {
  const accounts = await actions.withIdentity(clerkIdentity).listAccount();

  expect(accounts).toBeDefined();

});

test("list accounts by supervisor - authorised", async () => {

   const accounts = await actions.withIdentity(supervisorIdentity).listAccount();

    expect(accounts).toBeDefined();
});


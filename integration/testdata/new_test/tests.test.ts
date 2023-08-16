import { models, actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("two permissions rules, one includes fan out array", async () => {
  const identity = await models.identity.create({
    email: "adam@keel.xyz",
    password: "foo",
  });
  const user = await models.user.create({
    name: "Adam",
    identityId: identity.id,
  });
console.log(user);

const u = await models.user.findOne({ id: user.id });
console.log(u);
  await expect(
    actions.withIdentity(identity).getUser({
      id: user.id,
    })
  ).not.toHaveAuthorizationError();
});

// test("two permissions rules, one includes fan out array ---", async () => {
//   const orgKeel = await models.organisation.create({ name: "Keel"});
  
//   const identityAdam = await models.identity.create({
//     email: "adam@keel.xyz",
//     password: "foo",
//   });
//   const userAdam = await models.user.create({
//     name: "Adam",
//     identityId: identityAdam.id,
//   });
//   await models.userOrganisation.create({ userId: userAdam.id, organisationId: orgKeel.id });

//   const identityDave = await models.identity.create({
//     email: "dave@keel.xyz",
//     password: "foo",
//   });
//   const userDave = await models.user.create({
//     name: "Dave",
//     identityId: identityDave.id,
//   });
//   await models.userOrganisation.create({ userId: userDave.id, organisationId: orgKeel.id });

//   await expect(
//     actions.withIdentity(identityAdam).getUser({
//       id: userAdam.id,
//     })
//   ).not.toHaveAuthorizationError();

//   await expect(
//     actions.withIdentity(identityDave).getUser({
//       id: userAdam.id,
//     })
//   ).not.toHaveAuthorizationError();
// });

import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("complex role-based user model - with built-in actions", async () => {
  const identityDave = await models.identity.create({ email: "dave@keel.xyz" });
  const userDave = await models.user.create({
    name: "Dave",
    identityId: identityDave.id,
  });

  const identityAdam = await models.identity.create({ email: "adam@keel.xyz" });
  const userAdam = await models.user.create({
    name: "Adam",
    identityId: identityAdam.id,
  });

  const identityPete = await models.identity.create({ email: "pete@keel.xyz" });
  const userPete = await models.user.create({
    name: "Adam",
    identityId: identityAdam.id,
  });

  const identityNouser = await models.identity.create({
    email: "nouser@keel.xyz",
  });

  const projectWeave = await models.project.create({ name: "Weave" });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectWeave.id,
    role: "Admin",
  });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectWeave.id,
    role: "Developer",
  });
  await models.projectUser.create({
    userId: userAdam.id,
    projectId: projectWeave.id,
    role: "Developer",
  });

  const projectKeel = await models.project.create({ name: "Keel" });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectKeel.id,
    role: "Developer",
  });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectKeel.id,
    role: "Admin",
  });
  await models.projectUser.create({
    userId: userAdam.id,
    projectId: projectKeel.id,
    role: "Admin",
  });

  const projectPeteCo = await models.project.create({ name: "PeteCo" });
  await models.projectUser.create({
    userId: userPete.id,
    projectId: projectPeteCo.id,
    role: "Developer",
  });

  const projectTaskless = await models.project.create({ name: "Taskless" });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectTaskless.id,
    role: "Admin",
  });
  await models.projectUser.create({
    userId: userAdam.id,
    projectId: projectTaskless.id,
    role: "Admin",
  });
  await models.projectUser.create({
    userId: userPete.id,
    projectId: projectTaskless.id,
    role: "Admin",
  });

  const projectUserless = await models.project.create({ name: "Taskless" });

  const weaveTask = await models.task.create({
    title: "Weave task 1",
    projectId: projectWeave.id,
  });
  await models.task.create({
    title: "Weave task 2",
    projectId: projectWeave.id,
  });
  const keelTask = await models.task.create({
    title: "Keel task 1",
    projectId: projectKeel.id,
  });
  await models.task.create({ title: "Keel task 2", projectId: projectKeel.id });
  const peteCoTask = await models.task.create({
    title: "Pete Co task 1",
    projectId: projectPeteCo.id,
  });
  await models.task.create({ title: "Pete Co 2", projectId: projectPeteCo.id });
  const userlessTask = await models.task.create({
    title: "Userless task 1",
    projectId: projectUserless.id,
  });
  await models.task.create({
    title: "Userless task 2",
    projectId: projectUserless.id,
  });

  // Has one admin, tests that a project user who is admin of another project is not permitted.
  await expect(
    actions.withIdentity(identityDave).getTask({ id: weaveTask.id })
  ).resolves.toMatchObject({ id: weaveTask.id });
  await expect(
    actions.withIdentity(identityAdam).getTask({ id: weaveTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityPete).getTask({ id: weaveTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityNouser).getTask({ id: weaveTask.id })
  ).toHaveAuthorizationError();

  // Has two admins, tests that admins of other projects aren't permitted.
  await expect(
    actions.withIdentity(identityDave).getTask({ id: keelTask.id })
  ).resolves.toMatchObject({ id: keelTask.id });
  await expect(
    actions.withIdentity(identityAdam).getTask({ id: keelTask.id })
  ).resolves.toMatchObject({ id: keelTask.id });
  await expect(
    actions.withIdentity(identityPete).getTask({ id: keelTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityNouser).getTask({ id: keelTask.id })
  ).toHaveAuthorizationError();

  // Has only non-admin users, therefore no-one is permitted.
  await expect(
    actions.withIdentity(identityDave).getTask({ id: peteCoTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityAdam).getTask({ id: peteCoTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityPete).getTask({ id: peteCoTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityNouser).getTask({ id: peteCoTask.id })
  ).toHaveAuthorizationError();

  // // Has no users at all, therefore no-one is permitted.
  await expect(
    actions.withIdentity(identityDave).getTask({ id: userlessTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityAdam).getTask({ id: userlessTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityPete).getTask({ id: userlessTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityNouser).getTask({ id: userlessTask.id })
  ).toHaveAuthorizationError();
});

test("complex role-based user model - with functions", async () => {
  const identityDave = await models.identity.create({ email: "dave@keel.xyz" });
  const userDave = await models.user.create({
    name: "Dave",
    identityId: identityDave.id,
  });

  const identityAdam = await models.identity.create({ email: "adam@keel.xyz" });
  const userAdam = await models.user.create({
    name: "Adam",
    identityId: identityAdam.id,
  });

  const identityPete = await models.identity.create({ email: "pete@keel.xyz" });
  const userPete = await models.user.create({
    name: "Adam",
    identityId: identityAdam.id,
  });

  const identityNouser = await models.identity.create({
    email: "nouser@keel.xyz",
  });

  const projectWeave = await models.project.create({ name: "Weave" });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectWeave.id,
    role: "Admin",
  });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectWeave.id,
    role: "Developer",
  });
  await models.projectUser.create({
    userId: userAdam.id,
    projectId: projectWeave.id,
    role: "Developer",
  });

  const projectKeel = await models.project.create({ name: "Keel" });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectKeel.id,
    role: "Developer",
  });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectKeel.id,
    role: "Admin",
  });
  await models.projectUser.create({
    userId: userAdam.id,
    projectId: projectKeel.id,
    role: "Admin",
  });

  const projectPeteCo = await models.project.create({ name: "PeteCo" });
  await models.projectUser.create({
    userId: userPete.id,
    projectId: projectPeteCo.id,
    role: "Developer",
  });

  const projectTaskless = await models.project.create({ name: "Taskless" });
  await models.projectUser.create({
    userId: userDave.id,
    projectId: projectTaskless.id,
    role: "Admin",
  });
  await models.projectUser.create({
    userId: userAdam.id,
    projectId: projectTaskless.id,
    role: "Admin",
  });
  await models.projectUser.create({
    userId: userPete.id,
    projectId: projectTaskless.id,
    role: "Admin",
  });

  const projectUserless = await models.project.create({ name: "Taskless" });

  const weaveTask = await models.task.create({
    title: "Weave task 1",
    projectId: projectWeave.id,
  });
  await models.task.create({
    title: "Weave task 2",
    projectId: projectWeave.id,
  });
  const keelTask = await models.task.create({
    title: "Keel task 1",
    projectId: projectKeel.id,
  });
  await models.task.create({ title: "Keel task 2", projectId: projectKeel.id });
  const peteCoTask = await models.task.create({
    title: "Pete Co task 1",
    projectId: projectPeteCo.id,
  });
  await models.task.create({ title: "Pete Co 2", projectId: projectPeteCo.id });
  const userlessTask = await models.task.create({
    title: "Userless task 1",
    projectId: projectUserless.id,
  });
  await models.task.create({
    title: "Userless task 2",
    projectId: projectUserless.id,
  });

  // Has one admin, tests that a project user who is admin of another project is not permitted.
  await expect(
    actions.withIdentity(identityDave).getTaskFn({ id: weaveTask.id })
  ).resolves.toMatchObject({ id: weaveTask.id });
  await expect(
    actions.withIdentity(identityAdam).getTaskFn({ id: weaveTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityPete).getTaskFn({ id: weaveTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityNouser).getTaskFn({ id: weaveTask.id })
  ).toHaveAuthorizationError();

  // Has two admins, tests that admins of other projects aren't permitted.
  await expect(
    actions.withIdentity(identityDave).getTaskFn({ id: keelTask.id })
  ).resolves.toMatchObject({ id: keelTask.id });
  await expect(
    actions.withIdentity(identityAdam).getTaskFn({ id: keelTask.id })
  ).resolves.toMatchObject({ id: keelTask.id });
  await expect(
    actions.withIdentity(identityPete).getTaskFn({ id: keelTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityNouser).getTaskFn({ id: keelTask.id })
  ).toHaveAuthorizationError();

  // Has only non-admin users, therefore no-one is permitted.
  await expect(
    actions.withIdentity(identityDave).getTaskFn({ id: peteCoTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityAdam).getTaskFn({ id: peteCoTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityPete).getTaskFn({ id: peteCoTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityNouser).getTaskFn({ id: peteCoTask.id })
  ).toHaveAuthorizationError();

  // // Has no users at all, therefore no-one is permitted.
  await expect(
    actions.withIdentity(identityDave).getTaskFn({ id: userlessTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityAdam).getTaskFn({ id: userlessTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityPete).getTaskFn({ id: userlessTask.id })
  ).toHaveAuthorizationError();
  await expect(
    actions.withIdentity(identityNouser).getTaskFn({ id: userlessTask.id })
  ).toHaveAuthorizationError();
});

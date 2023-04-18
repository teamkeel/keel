import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("empty project id", async () => {
  const identity = await models.identity.create({
    email: "adam@keel.xyz",
    password: "123",
  });
  const project = await models.project.create({
    name: "my project",
  });
  await models.todo.create({
    projectId: project.id,
    label: "my todo",
    ownerId: identity.id,
  });

  const { results } = await actions
    .withIdentity(identity)
    .listTodo({ where: { projectId: { oneOf: [] } } });

  expect(results.map((r) => r.id)).toEqual([]);
});

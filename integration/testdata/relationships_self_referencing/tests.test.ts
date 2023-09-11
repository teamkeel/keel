import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("self referencing model - one to many", async () => {
  const grandmother = await actions.createPerson({ name: "Grandmother" });

  const keelson = await actions.createPerson({
    name: "Mrs Keelson",
    mother: { id: grandmother.id },
  });
  const keelam = await actions.createPerson({
    name: "Keelam",
    mother: { id: keelson.id },
  });
  const keelya = await actions.createPerson({
    name: "Keeyla",
    mother: { id: keelson.id },
  });

  const weaveton = await actions.createPerson({ name: "Mrs Weaveton" });
  const woolam = await actions.createPerson({
    name: "Woolam",
    mother: { id: weaveton.id },
  });
  const wayla = await actions.createPerson({
    name: "Wayla",
    mother: { id: weaveton.id },
  });

  const grandmotherChildren = await actions.children({
    where: { mother: { id: { equals: grandmother.id } } },
  });
  expect(grandmotherChildren.results).toHaveLength(1);
  expect(grandmotherChildren.results[0].id).toEqual(keelson.id);

  const keelsonChildren = await actions.children({
    where: { mother: { id: { equals: keelson.id } } },
  });
  expect(keelsonChildren.results).toHaveLength(2);
  expect(keelsonChildren.results[0].id).toEqual(keelam.id);
  expect(keelsonChildren.results[1].id).toEqual(keelya.id);

  const weavetonChildren = await actions.children({
    where: { mother: { id: { equals: weaveton.id } } },
  });
  expect(weavetonChildren.results).toHaveLength(2);
  expect(weavetonChildren.results[0].id).toEqual(wayla.id);
  expect(weavetonChildren.results[1].id).toEqual(woolam.id);

  const weavetonChildrenByName = await actions.children({
    where: { mother: { name: { equals: "Mrs Weaveton" } } },
  });
  expect(weavetonChildrenByName.results).toHaveLength(2);
  expect(weavetonChildrenByName.results[0].id).toEqual(wayla.id);
  expect(weavetonChildrenByName.results[1].id).toEqual(woolam.id);

  const waylaChildrenByName = await actions.children({
    where: { mother: { name: { equals: wayla.name } } },
  });
  expect(waylaChildrenByName.results).toHaveLength(0);
});

test("self referencing model - many to one", async () => {
  const grandmother = await actions.createPerson({ name: "Grandmother" });

  const keelson = await actions.createPerson({
    name: "Mrs Keelson",
    mother: { id: grandmother.id },
  });
  const keelam = await actions.createPerson({
    name: "Keelam",
    mother: { id: keelson.id },
  });
  const keelya = await actions.createPerson({
    name: "Keeyla",
    mother: { id: keelson.id },
  });

  const weaveton = await actions.createPerson({ name: "Mrs Weaveton" });
  const woolam = await actions.createPerson({
    name: "Woolam",
    mother: { id: weaveton.id },
  });
  const wayla = await actions.createPerson({
    name: "Wayla",
    mother: { id: weaveton.id },
  });

  const motherofWoolam = await actions.mothersOf({
    where: {
      children: { id: { equals: woolam.id } },
    },
  });
  expect(motherofWoolam.results).toHaveLength(1);
  expect(motherofWoolam.results[0].id).toEqual(weaveton.id);

  const mothersOfWoolamAndKeelson = await actions.mothersOf({
    where: { children: { id: { oneOf: [woolam.id, keelson.id] } } },
  });
  expect(mothersOfWoolamAndKeelson.results).toHaveLength(2);
  expect(mothersOfWoolamAndKeelson.results[0].id).toEqual(grandmother.id);
  expect(mothersOfWoolamAndKeelson.results[1].id).toEqual(weaveton.id);

  const motherofKeelamAndKeelyaByName = await actions.mothersOf({
    where: {
      children: { name: { oneOf: [keelam.name, keelya.name] } },
    },
  });
  expect(motherofKeelamAndKeelyaByName.results).toHaveLength(1);
  expect(motherofKeelamAndKeelyaByName.results[0].id).toEqual(keelson.id);
});

import { test, expect } from "vitest";
import { PrismaClient } from "@prisma/client";

test("prisma client", async () => {
  const p = new PrismaClient();

  const person = await p.person.create({
    data: {
      id: "very-unique-value",
      createdAt: new Date(),
      updatedAt: new Date(),
      name: "foo",
    },
  });

  expect(person.name).toBe("foo");
});

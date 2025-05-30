import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase, actions } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("computed fields - self referencing", async () => {
  const car = await actions.createCar();

  const mileage1 = await actions.createMileage({
    car: { id: car.id },
    miles: 100,
    date: new Date(),
    previous: null,
  });

  expect(mileage1.diffFromPrevious).toBe(null);

  const mileage2 = await actions.createMileage({
    car: { id: car.id },
    miles: 155,
    date: new Date(),
    previous: { id: mileage1.id },
  });

  expect(mileage2.diffFromPrevious).toBe(55);
});

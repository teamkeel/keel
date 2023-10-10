import { actions, models, resetDatabase } from "@teamkeel/testing";
import { MealName } from "@teamkeel/sdk";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("create relationships - many to many", async () => {
  const identity = await models.identity.create({ email: "bobbi@keel.so" });

  const bob = await models.person.create({ name: "Bob", email: "bob@keel.so" });
  const mo = await models.person.create({ name: "Mo", email: "mo@keel.so" });
  const mary = await models.person.create({
    name: "Mary",
    email: "mary@keel.so",
  });

  const retreat = await actions.withIdentity(identity).createRetreat({
    retreatName: "Cape Town",
    attendees: [{ person: { id: bob.id } }, { person: { id: mo.id } }],
  });

  const retreatPersons = await models.retreatPerson.findMany({
    where: { retreat: { id: retreat.id } },
  });

  expect(retreatPersons).toHaveLength(2);
  expect(retreatPersons.filter((rp) => rp.personId == mo.id)).toHaveLength(1);
  expect(retreatPersons.filter((rp) => rp.personId == bob.id)).toHaveLength(1);
});

test("create relationships - many to many to many", async () => {
  const identity = await models.identity.create({ email: "bobbi@keel.so" });

  const retreat = await actions
    .withIdentity(identity)
    .createRetreatWithPeopleAndMeals({
      retreatName: "Cape Town",
      attendees: [
        {
          person: {
            name: "Bob",
            email: "bob@keel.so",
            meals: [
              {
                meal: {
                  mealName: MealName.Lunch,
                },
              },
            ],
          },
        },
        {
          person: {
            name: "Mo",
            email: "mo@keel.so",
            meals: [
              {
                meal: {
                  mealName: MealName.Breakfast,
                },
              },
              {
                meal: {
                  mealName: MealName.Dinner,
                },
              },
            ],
          },
        },
      ],
    });

  const retreatPersons = await models.retreatPerson.findMany({
    where: { retreat: { id: retreat.id } },
  });
  expect(retreatPersons).toHaveLength(2);

  const mo = await models.person.findOne({ email: "mo@keel.so" });
  expect(mo?.name).toEqual("Mo");

  const mealPersons = await models.mealPerson.findMany();
  expect(mealPersons).toHaveLength(3);

  const moMealpersons = mealPersons.filter((mp) => mp.personId == mo?.id);
  expect(moMealpersons).toHaveLength(2);

  const meals = await models.meal.findMany();
  expect(meals).toHaveLength(3);

  const moMeals = meals.filter(
    (meal) =>
      meal.id == moMealpersons[0].mealId || meal.id == moMealpersons[1].mealId
  );

  expect(moMeals.filter((m) => m.mealName == MealName.Breakfast)).toHaveLength(
    1
  );
  expect(moMeals.filter((m) => m.mealName == MealName.Dinner)).toHaveLength(1);
});

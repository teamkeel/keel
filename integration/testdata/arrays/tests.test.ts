import { actions, models, resetDatabase } from "@teamkeel/testing";
import { MyEnum } from "@teamkeel/sdk";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("array fields - create action", async () => {
  const thing = await actions.createThing({
    texts: ["Keel", "Weave"],
    numbers: [1, 2, 3],
    booleans: [true, true, false],
    dates: [
      new Date("2023-01-02T00:00:00.123+00:00"),
      new Date("2024-12-31Z+00:00"),
      new Date("2025-07-03T23:59:59+00:00"),
    ],
    timestamps: [
      new Date("2023-01-02 23:00:30"),
      new Date("2023-11-13 06:17:30.123"),
      new Date("2024-02-01 23:00:30"),
    ],
    enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
  });

  expect(thing.texts).toHaveLength(2);
  expect(thing.texts![0]).toEqual("Keel");
  expect(thing.texts![1]).toEqual("Weave");

  expect(thing.numbers).toHaveLength(3);
  expect(thing.numbers![0]).toEqual(1);
  expect(thing.numbers![1]).toEqual(2);
  expect(thing.numbers![2]).toEqual(3);

  expect(thing.booleans).toHaveLength(3);
  expect(thing.booleans![0]).toEqual(true);
  expect(thing.booleans![1]).toEqual(true);
  expect(thing.booleans![2]).toEqual(false);

  expect(thing.dates).toHaveLength(3);
  expect(thing.dates![0]).toEqual(new Date("2023-01-02T00:00:00.000+00:00"));
  expect(thing.dates![1]).toEqual(new Date("2024-12-31T00:00:00.000+00:00"));
  expect(thing.dates![2]).toEqual(new Date("2025-07-03T00:00:00+00:00"));

  expect(thing.timestamps).toHaveLength(3);
  expect(thing.timestamps![0]).toEqual(new Date("2023-01-02 23:00:30"));
  expect(thing.timestamps![1]).toEqual(new Date("2023-11-13 06:17:30.123"));
  expect(thing.timestamps![2]).toEqual(new Date("2024-02-01 23:00:30.000"));

  expect(thing.enums).toHaveLength(3);
  expect(thing.enums![0]).toEqual(MyEnum.One);
  expect(thing.enums![1]).toEqual(MyEnum.Two);
  expect(thing.enums![2]).toEqual(MyEnum.Three);
});

test("array fields - empty arrays", async () => {
  const thing = await actions.createThing({
    texts: [],
    numbers: [],
    booleans: [],
    dates: [],
    timestamps: [],
    enums: [],
  });

  expect(thing.texts).not.toBeNull();
  expect(thing.texts).toHaveLength(0);

  expect(thing.numbers).not.toBeNull();
  expect(thing.numbers).toHaveLength(0);

  expect(thing.booleans).not.toBeNull();
  expect(thing.booleans).toHaveLength(0);

  expect(thing.dates).not.toBeNull();
  expect(thing.dates).toHaveLength(0);

  expect(thing.timestamps).not.toBeNull();
  expect(thing.timestamps).toHaveLength(0);

  expect(thing.enums).not.toBeNull();
  expect(thing.enums).toHaveLength(0);
});

test("array fields - null arrays", async () => {
  const thing = await actions.createThing({
    texts: null,
    numbers: null,
    booleans: null,
    dates: null,
    timestamps: null,
    enums: null,
  });

  expect(thing.texts).toBeNull();
  expect(thing.numbers).toBeNull();
  expect(thing.booleans).toBeNull();
  expect(thing.dates).toBeNull();
  expect(thing.timestamps).toBeNull();
  expect(thing.enums).toBeNull();
});

test("array fields - update action", async () => {
  const created = await actions.createThing({
    texts: ["nope"],
    numbers: [101, 102, 103],
    booleans: [false, false, true, true],
    dates: [new Date("1999-01-02")],
    timestamps: [new Date("2023-01-02 23:00:30")],
    enums: [MyEnum.Three],
  });

  const thing = await actions.updateThing({
    where: { id: created.id },
    values: {
      texts: ["Keel", "Weave"],
      numbers: [1, 2, 3],
      booleans: [true, true, false],
      dates: [
        new Date("2023-01-02T00:00:00.123+00:00"),
        new Date("2024-12-31Z+00:00"),
        new Date("2025-07-03T23:59:59+00:00"),
      ],
      timestamps: [
        new Date("2023-01-02 23:00:30"),
        new Date("2023-11-13 06:17:30.123"),
        new Date("2024-02-01 23:00:30"),
      ],
      enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
    },
  });

  expect(thing.texts).toHaveLength(2);
  expect(thing.texts![0]).toEqual("Keel");
  expect(thing.texts![1]).toEqual("Weave");

  expect(thing.numbers).toHaveLength(3);
  expect(thing.numbers![0]).toEqual(1);
  expect(thing.numbers![1]).toEqual(2);
  expect(thing.numbers![2]).toEqual(3);

  expect(thing.booleans).toHaveLength(3);
  expect(thing.booleans![0]).toEqual(true);
  expect(thing.booleans![1]).toEqual(true);
  expect(thing.booleans![2]).toEqual(false);

  expect(thing.dates).toHaveLength(3);
  expect(thing.dates![0]).toEqual(new Date("2023-01-02T00:00:00.000+00:00"));
  expect(thing.dates![1]).toEqual(new Date("2024-12-31T00:00:00.000+00:00"));
  expect(thing.dates![2]).toEqual(new Date("2025-07-03T00:00:00+00:00"));

  expect(thing.timestamps).toHaveLength(3);
  expect(thing.timestamps![0]).toEqual(new Date("2023-01-02 23:00:30"));
  expect(thing.timestamps![1]).toEqual(new Date("2023-11-13 06:17:30.123"));
  expect(thing.timestamps![2]).toEqual(new Date("2024-02-01 23:00:30.000"));

  expect(thing.enums).toHaveLength(3);
  expect(thing.enums![0]).toEqual(MyEnum.One);
  expect(thing.enums![1]).toEqual(MyEnum.Two);
  expect(thing.enums![2]).toEqual(MyEnum.Three);
});

test("array fields - get action", async () => {
  const created = await actions.createThing({
    texts: ["Keel", "Weave"],
    numbers: [1, 2, 3],
    booleans: [true, true, false],
    dates: [
      new Date("2023-01-02T00:00:00.123+00:00"),
      new Date("2024-12-31Z+00:00"),
      new Date("2025-07-03T23:59:59+00:00"),
    ],
    timestamps: [
      new Date("2023-01-02 23:00:30"),
      new Date("2023-11-13 06:17:30.123"),
      new Date("2024-02-01 23:00:30"),
    ],
    enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
  });

  const thing = await actions.getThing({
    id: created.id,
  });

  expect(thing?.texts).toHaveLength(2);
  expect(thing?.texts![0]).toEqual("Keel");
  expect(thing?.texts![1]).toEqual("Weave");

  expect(thing?.numbers).toHaveLength(3);
  expect(thing?.numbers![0]).toEqual(1);
  expect(thing?.numbers![1]).toEqual(2);
  expect(thing?.numbers![2]).toEqual(3);

  expect(thing?.booleans).toHaveLength(3);
  expect(thing?.booleans![0]).toEqual(true);
  expect(thing?.booleans![1]).toEqual(true);
  expect(thing?.booleans![2]).toEqual(false);

  expect(thing?.dates).toHaveLength(3);
  expect(thing?.dates![0]).toEqual(new Date("2023-01-02T00:00:00.000+00:00"));
  expect(thing?.dates![1]).toEqual(new Date("2024-12-31T00:00:00.000+00:00"));
  expect(thing?.dates![2]).toEqual(new Date("2025-07-03T00:00:00+00:00"));

  expect(thing?.timestamps).toHaveLength(3);
  expect(thing?.timestamps![0]).toEqual(new Date("2023-01-02 23:00:30"));
  expect(thing?.timestamps![1]).toEqual(new Date("2023-11-13 06:17:30.123"));
  expect(thing?.timestamps![2]).toEqual(new Date("2024-02-01 23:00:30.000"));

  expect(thing?.enums).toHaveLength(3);
  expect(thing?.enums![0]).toEqual(MyEnum.One);
  expect(thing?.enums![1]).toEqual(MyEnum.Two);
  expect(thing?.enums![2]).toEqual(MyEnum.Three);
});

test("array fields - list action", async () => {
  await actions.createThing({
    texts: ["Keel", "Weave"],
    numbers: [1, 2, 3],
    booleans: [true, true, false],
    dates: [
      new Date("2023-01-02T00:00:00.123+00:00"),
      new Date("2024-12-31Z+00:00"),
      new Date("2025-07-03T23:59:59+00:00"),
    ],
    timestamps: [
      new Date("2023-01-02 23:00:30"),
      new Date("2023-11-13 06:17:30.123"),
      new Date("2024-02-01 23:00:30"),
    ],
    enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
  });

  const things = await actions.listThings();

  expect(things.results).toHaveLength(1);

  const thing = things.results[0];

  expect(thing.texts).toHaveLength(2);
  expect(thing.texts![0]).toEqual("Keel");
  expect(thing.texts![1]).toEqual("Weave");

  expect(thing.numbers).toHaveLength(3);
  expect(thing.numbers![0]).toEqual(1);
  expect(thing.numbers![1]).toEqual(2);
  expect(thing.numbers![2]).toEqual(3);

  expect(thing.booleans).toHaveLength(3);
  expect(thing.booleans![0]).toEqual(true);
  expect(thing.booleans![1]).toEqual(true);
  expect(thing.booleans![2]).toEqual(false);

  expect(thing.dates).toHaveLength(3);
  expect(thing.dates![0]).toEqual(new Date("2023-01-02T00:00:00.000+00:00"));
  expect(thing.dates![1]).toEqual(new Date("2024-12-31T00:00:00.000+00:00"));
  expect(thing.dates![2]).toEqual(new Date("2025-07-03T00:00:00+00:00"));

  expect(thing.timestamps).toHaveLength(3);
  expect(thing.timestamps![0]).toEqual(new Date("2023-01-02 23:00:30"));
  expect(thing.timestamps![1]).toEqual(new Date("2023-11-13 06:17:30.123"));
  expect(thing.timestamps![2]).toEqual(new Date("2024-02-01 23:00:30.000"));

  expect(thing.enums).toHaveLength(3);
  expect(thing.enums![0]).toEqual(MyEnum.One);
  expect(thing.enums![1]).toEqual(MyEnum.Two);
  expect(thing.enums![2]).toEqual(MyEnum.Three);
});

test("array fields - list action implicit querying - text", async () => {
  const t1 = await actions.createThing({
    texts: ["Keel", "Weave"],
  });

  const t2 = await actions.createThing({
    texts: ["Keel", "Weave", "Keelson", "Keeler"],
  });

  const t3 = await actions.createThing({
    texts: ["Keel", "Weave"],
  });

  const t4 = await actions.createThing({
    texts: null,
  });

  const t5 = await actions.createThing({
    texts: [],
  });

  const t6 = await actions.createThing({
    texts: ["Weave", "Keel"],
  });

  const t7 = await actions.createThing({
    texts: ["Keelson", "Keelson"],
  });

  const things1 = await actions.listThings({
    where: {
      texts: {
        equals: ["Keel", "Weave"],
      },
    },
  });

  expect(things1.results).toHaveLength(2);
  expect(things1.results[0].id).toEqual(t1.id);
  expect(things1.results[1].id).toEqual(t3.id);

  const things2 = await actions.listThings({
    where: {
      texts: {
        notEquals: ["Keel", "Weave"],
      },
    },
  });

  expect(things2.results).toHaveLength(5);
  expect(things2.results[0].id).toEqual(t2.id);
  expect(things2.results[1].id).toEqual(t4.id);
  expect(things2.results[2].id).toEqual(t5.id);
  expect(things2.results[3].id).toEqual(t6.id);
  expect(things2.results[4].id).toEqual(t7.id);

  const things3 = await actions.listThings({
    where: {
      texts: {
        equals: null,
      },
    },
  });

  expect(things3.results).toHaveLength(1);
  expect(things3.results[0].id).toEqual(t4.id);

  const things4 = await actions.listThings({
    where: {
      texts: {
        notEquals: null,
      },
    },
  });

  expect(things4.results).toHaveLength(6);
  expect(things4.results[0].id).toEqual(t1.id);
  expect(things4.results[1].id).toEqual(t2.id);
  expect(things4.results[2].id).toEqual(t3.id);
  expect(things4.results[3].id).toEqual(t5.id);
  expect(things4.results[4].id).toEqual(t6.id);
  expect(things4.results[5].id).toEqual(t7.id);

  const things5 = await actions.listThings({
    where: {
      texts: {
        equals: [],
      },
    },
  });

  expect(things5.results).toHaveLength(1);
  expect(things5.results[0].id).toEqual(t5.id);

  const things6 = await actions.listThings({
    where: {
      texts: {
        notEquals: [],
      },
    },
  });

  expect(things6.results).toHaveLength(6);
  expect(things6.results[0].id).toEqual(t1.id);
  expect(things6.results[1].id).toEqual(t2.id);
  expect(things6.results[2].id).toEqual(t3.id);
  expect(things6.results[3].id).toEqual(t4.id);
  expect(things6.results[4].id).toEqual(t6.id);
  expect(things6.results[5].id).toEqual(t7.id);

  const things7 = await actions.listThings({
    where: {
      texts: {
        any: {
          equals: "Weave"
        },
      },
    },
  });

  expect(things7.results).toHaveLength(4);
  expect(things7.results[0].id).toEqual(t1.id);
  expect(things7.results[1].id).toEqual(t2.id);
  expect(things7.results[2].id).toEqual(t3.id);
  expect(things7.results[3].id).toEqual(t6.id);

  const things8 = await actions.listThings({
    where: {
      texts: {
        all: {
          equals: "Keelson"
        },
      },
    },
  });

  expect(things8.results).toHaveLength(1);
  expect(things8.results[0].id).toEqual(t7.id);

  const things9 = await actions.listThings({
    where: {
      texts: {
        any: {
          equals: "Keelson",
          notEquals: "Weave"
        },
      },
    },
  });

  expect(things9.results).toHaveLength(1);
  expect(things9.results[0].id).toEqual(t7.id);

  const things10 = await actions.listThings({
    where: {
      texts: {
        any: {
          notEquals: "Weave"
        },
      },
    },
  });

  expect(things10.results).toHaveLength(3);
  expect(things10.results[0].id).toEqual(t4.id);
  expect(things10.results[1].id).toEqual(t5.id);
  expect(things10.results[2].id).toEqual(t7.id);

  const things11 = await actions.listThings({
    where: {
      texts: {
        all: {
          notEquals: "Keelson"
        },
      },
    },
  });
console.log(things11.results);
  expect(things11.results).toHaveLength(6);
  expect(things11.results[0].id).toEqual(t1.id);
  expect(things11.results[1].id).toEqual(t2.id);
  expect(things11.results[2].id).toEqual(t3.id);
  expect(things11.results[3].id).toEqual(t4.id);
  expect(things11.results[4].id).toEqual(t5.id);
  expect(things11.results[5].id).toEqual(t6.id);

});

test("array fields - list action implicit querying - number", async () => {
  const t1 = await actions.createThing({
    numbers: [1, 2],
  });

  const t2 = await actions.createThing({
    numbers: [1, 2, 3, 4],
  });

  const t3 = await actions.createThing({
    numbers: [1, 2],
  });

  const t4 = await actions.createThing({
    numbers: null,
  });

  const t5 = await actions.createThing({
    numbers: [],
  });

  const t6 = await actions.createThing({
    numbers: [2, 1],
  });

  const things = await actions.listThings({
    where: {
      numbers: {
        equals: [1, 2],
      },
    },
  });

  expect(things.results).toHaveLength(2);
  expect(things.results[0].id).toEqual(t1.id);
  expect(things.results[1].id).toEqual(t3.id);
});

test("array fields - list action implicit querying - date", async () => {
  const t1 = await actions.createThing({
    dates: [new Date(2024, 1, 1, 0, 0, 0, 0), new Date(2024, 1, 2, 0, 0, 0, 0)],
  });

  const t2 = await actions.createThing({
    dates: [
      new Date(2024, 1, 1, 0, 0, 0, 0),
      new Date(2024, 1, 2, 0, 0, 0, 0),
      new Date(2024, 1, 3, 0, 0, 0, 0),
    ],
  });

  const t3 = await actions.createThing({
    dates: [new Date(2024, 1, 1, 0, 0, 0, 0), new Date(2024, 1, 2, 0, 0, 0, 0)],
  });

  const t4 = await actions.createThing({
    dates: null,
  });

  const t5 = await actions.createThing({
    dates: [],
  });

  const t6 = await actions.createThing({
    dates: [new Date(2024, 1, 2, 0, 0, 0, 0), new Date(2024, 1, 1, 0, 0, 0, 0)],
  });

  const things = await actions.listThings({
    where: {
      dates: {
        equals: [
          new Date(2024, 1, 1, 0, 0, 0, 0),
          new Date(2024, 1, 2, 0, 0, 0, 0),
        ],
      },
    },
  });

  expect(things.results).toHaveLength(2);
  expect(things.results[0].id).toEqual(t1.id);
  expect(things.results[1].id).toEqual(t3.id);
});

test("array fields - list action implicit querying - timestamp", async () => {
  const t1 = await actions.createThing({
    timestamps: [
      new Date(2024, 1, 1, 30, 45, 50, 0),
      new Date(2024, 1, 2, 59, 0, 0, 0),
    ],
  });

  const t2 = await actions.createThing({
    timestamps: [
      new Date(2024, 1, 1, 30, 45, 50, 0),
      new Date(2024, 1, 2, 59, 0, 0, 0),
      new Date(2024, 1, 3, 0, 0, 0, 0),
    ],
  });

  const t3 = await actions.createThing({
    timestamps: [
      new Date(2024, 1, 1, 30, 45, 50, 0),
      new Date(2024, 1, 2, 59, 0, 0, 0),
    ],
  });

  const t4 = await actions.createThing({
    timestamps: null,
  });

  const t5 = await actions.createThing({
    timestamps: [],
  });

  const t6 = await actions.createThing({
    timestamps: [
      new Date(2024, 1, 2, 59, 0, 0, 0),
      new Date(2024, 1, 1, 30, 45, 50, 0),
    ],
  });

  const things = await actions.listThings({
    where: {
      timestamps: {
        equals: [
          new Date(2024, 1, 1, 30, 45, 50, 0),
          new Date(2024, 1, 2, 59, 0, 0, 0),
        ],
      },
    },
  });

  expect(things.results).toHaveLength(2);
  expect(things.results[0].id).toEqual(t1.id);
  expect(things.results[1].id).toEqual(t3.id);
});

test("array fields - list action implicit querying - enums", async () => {
  const t1 = await actions.createThing({
    enums: [MyEnum.One, MyEnum.Two],
  });

  const t2 = await actions.createThing({
    enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
  });

  const t3 = await actions.createThing({
    enums: [MyEnum.One, MyEnum.Two],
  });

  const t4 = await actions.createThing({
    enums: null,
  });

  const t5 = await actions.createThing({
    enums: [],
  });

  const t6 = await actions.createThing({
    enums: [MyEnum.Two, MyEnum.One],
  });

  const things = await actions.listThings({
    where: {
      enums: {
        equals: [MyEnum.One, MyEnum.Two],
      },
    },
  });

  expect(things.results).toHaveLength(2);
  expect(things.results[0].id).toEqual(t1.id);
  expect(things.results[1].id).toEqual(t3.id);
});

test("arrays - set attribute - create", async () => {
  const created = await actions.createSet();

  expect(created.texts).toHaveLength(2);
  expect(created.texts![0]).toEqual("Keel");
  expect(created.texts![1]).toEqual("Weave");

  expect(created.numbers).toHaveLength(3);
  expect(created.numbers![0]).toEqual(1);
  expect(created.numbers![1]).toEqual(2);
  expect(created.numbers![2]).toEqual(3);

  expect(created.booleans).toHaveLength(3);
  expect(created.booleans![0]).toEqual(true);
  expect(created.booleans![1]).toEqual(true);
  expect(created.booleans![2]).toEqual(false);

  expect(created.dates).toBeNull();
  expect(created.timestamps).toBeNull();

  expect(created.enums).toHaveLength(3);
  expect(created.enums![0]).toEqual(MyEnum.One);
  expect(created.enums![1]).toEqual(MyEnum.Two);
  expect(created.enums![2]).toEqual(MyEnum.Three);
});

test("arrays - set attribute with empty arrays", async () => {
  const thing = await actions.createThing({
    texts: ["Keel", "Weave"],
    numbers: [1, 2, 3],
    booleans: [true, true, false],
    dates: [
      new Date("2023-01-02 00:00:00.123"),
      new Date("2024-12-31"),
      new Date("2025-07-03 23:59:59"),
    ],
    timestamps: [
      new Date("2023-01-02 23:00:30"),
      new Date("2023-11-13 06:17:30.123"),
      new Date("2024-02-01 23:00:30"),
    ],
    enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
  });

  const created = await actions.updateSetToEmpty({ where: { id: thing.id } });

  expect(created.texts).toHaveLength(0);
  expect(created.numbers).toHaveLength(0);
  expect(created.booleans).toHaveLength(0);
  expect(created.dates).toHaveLength(0);
  expect(created.timestamps).toHaveLength(0);
  expect(created.enums).toHaveLength(0);
});

test("arrays - set attribute with null", async () => {
  const thing = await actions.createThing({
    texts: ["Keel", "Weave"],
    numbers: [1, 2, 3],
    booleans: [true, true, false],
    dates: [
      new Date("2023-01-02 00:00:00.123"),
      new Date("2024-12-31"),
      new Date("2025-07-03 23:59:59"),
    ],
    timestamps: [
      new Date("2023-01-02 23:00:30"),
      new Date("2023-11-13 06:17:30.123"),
      new Date("2024-02-01 23:00:30"),
    ],
    enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
  });

  const created = await actions.updateSetToNull({ where: { id: thing.id } });

  expect(created.texts).toBeNull();
  expect(created.numbers).toBeNull();
  expect(created.booleans).toBeNull();
  expect(created.dates).toBeNull();
  expect(created.timestamps).toBeNull();
  expect(created.enums).toBeNull();
});

test("array fields - model api - create", async () => {
  const created = await models.thing.create({
    texts: ["Keel", "Weave"],
    numbers: [1, 2, 3],
    booleans: [true, true, false],
    dates: [
      new Date(2023, 1, 2, 0, 0, 0, 0),
      new Date(2024, 31, 12, 12, 45, 0, 0),
      new Date(2025, 13, 3, 0, 0, 0, 0),
    ],
    timestamps: [
      new Date("2023-01-02 23:00:30"),
      new Date("2023-11-13 06:17:30.123"),
      new Date("2024-02-01 23:00:30"),
    ],
    enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
  });

  const thing = await models.thing.findOne({ id: created.id });
  expect(thing?.texts).toEqual(["Keel", "Weave"]);
  expect(thing?.numbers).toEqual([1, 2, 3]);
  expect(thing?.booleans).toEqual([true, true, false]);
  expect(thing?.dates).toEqual([
    new Date(2023, 1, 2, 0, 0, 0, 0),
    new Date(2024, 31, 12, 0, 0, 0, 0),
    new Date(2025, 13, 3, 0, 0, 0, 0),
  ]);
  expect(thing?.timestamps).toEqual([
    new Date("2023-01-02 23:00:30"),
    new Date("2023-11-13 06:17:30.123"),
    new Date("2024-02-01 23:00:30"),
  ]);
  expect(thing?.enums).toEqual([MyEnum.One, MyEnum.Two, MyEnum.Three]);
});



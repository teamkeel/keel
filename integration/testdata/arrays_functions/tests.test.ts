import { models, resetDatabase, actions } from "@teamkeel/testing";
import { MyEnum, InlineFile, Duration } from "@teamkeel/sdk";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

// test("array functions - modelapi - create", async () => {
//   const created = await models.thing.create({
//     texts: ["Keel", "Weave"],
//     numbers: [1, 2, 3],
//     booleans: [true, true, false],
//     dates: [
//       new Date(2023, 1, 2, 0, 0, 0, 0),
//       new Date(2024, 31, 12, 12, 45, 0, 0),
//       new Date(2025, 13, 3, 0, 0, 0, 0),
//     ],
//     timestamps: [
//       new Date("2023-01-02 23:00:30"),
//       new Date("2023-11-13 06:17:30.123"),
//       new Date("2024-02-01 23:00:30"),
//     ],
//     enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
//     files: [
//       InlineFile.fromDataURL("data:text/plain;name=one.txt;base64,b25l=="),
//       InlineFile.fromDataURL("data:text/plain;name=two.txt;base64,dHdv=="),
//     ],
//     durations: [
//       Duration.fromISOString("PT2H3M4S"),
//       Duration.fromISOString("PT1H2M3S"),
//     ],
//   });

//   const thing = await models.thing.findOne({ id: created.id });
//   expect(thing?.texts).toEqual(["Keel", "Weave"]);
//   expect(thing?.numbers).toEqual([1, 2, 3]);
//   expect(thing?.booleans).toEqual([true, true, false]);
//   expect(thing?.dates).toEqual([
//     new Date(2023, 1, 2, 0, 0, 0, 0),
//     new Date(2024, 31, 12, 0, 0, 0, 0),
//     new Date(2025, 13, 3, 0, 0, 0, 0),
//   ]);
//   expect(thing?.timestamps).toEqual([
//     new Date("2023-01-02 23:00:30"),
//     new Date("2023-11-13 06:17:30.123"),
//     new Date("2024-02-01 23:00:30"),
//   ]);
//   expect(thing?.enums).toEqual([MyEnum.One, MyEnum.Two, MyEnum.Three]);

//   expect(thing?.files).toHaveLength(2);

//   expect(thing?.files?.[0].contentType).toEqual("text/plain");
//   expect(thing?.files?.[0].filename).toEqual("one.txt");
//   expect(thing?.files?.[0].size).toEqual(3);
//   const contents1 = await thing?.files?.[0].read();
//   expect(contents1?.toString("utf-8")).toEqual("one");

//   expect(thing?.files?.[1].contentType).toEqual("text/plain");
//   expect(thing?.files?.[1].filename).toEqual("two.txt");
//   expect(thing?.files?.[1].size).toEqual(3);
//   const contents2 = await thing?.files?.[1].read();
//   expect(contents2?.toString("utf-8")).toEqual("two");

//   expect(thing?.durations).toEqual([
//     {
//       hours: 2,
//       minutes: 3,
//       seconds: 4,
//     },
//     {
//       hours: 1,
//       minutes: 2,
//       seconds: 3,
//     },
//   ]);
// });

// test("array functions - modelapi - empty arrays", async () => {
//   const thing = await models.thing.create({
//     texts: [],
//     numbers: [],
//     booleans: [],
//     dates: [],
//     timestamps: [],
//     enums: [],
//     files: [],
//     durations: [],
//   });

//   expect(thing.texts).not.toBeNull();
//   expect(thing.texts).toHaveLength(0);

//   expect(thing.numbers).not.toBeNull();
//   expect(thing.numbers).toHaveLength(0);

//   expect(thing.booleans).not.toBeNull();
//   expect(thing.booleans).toHaveLength(0);

//   expect(thing.dates).not.toBeNull();
//   expect(thing.dates).toHaveLength(0);

//   expect(thing.timestamps).not.toBeNull();
//   expect(thing.timestamps).toHaveLength(0);

//   expect(thing.enums).not.toBeNull();
//   expect(thing.enums).toHaveLength(0);

//   expect(thing.files).not.toBeNull();
//   expect(thing.files).toHaveLength(0);

//   expect(thing.durations).not.toBeNull();
//   expect(thing.durations).toHaveLength(0);
// });

// test("array functions - null arrays", async () => {
//   const thing = await models.thing.create({
//     texts: null,
//     numbers: null,
//     booleans: null,
//     dates: null,
//     timestamps: null,
//     enums: null,
//     files: null,
//     durations: null,
//   });

//   expect(thing.texts).toBeNull();
//   expect(thing.numbers).toBeNull();
//   expect(thing.booleans).toBeNull();
//   expect(thing.dates).toBeNull();
//   expect(thing.timestamps).toBeNull();
//   expect(thing.enums).toBeNull();
//   expect(thing.files).toBeNull();
//   expect(thing.durations).toBeNull();
// });

// test("array functions - update action", async () => {
//   const created = await models.thing.create({
//     texts: ["Keel", "Weave"],
//     numbers: [1, 2, 3],
//     booleans: [true, true, false],
//     dates: [
//       new Date(2023, 1, 2, 0, 0, 0, 0),
//       new Date(2024, 31, 12, 12, 45, 0, 0),
//       new Date(2025, 13, 3, 0, 0, 0, 0),
//     ],
//     timestamps: [
//       new Date("2023-01-02 23:00:30"),
//       new Date("2023-11-13 06:17:30.123"),
//       new Date("2024-02-01 23:00:30"),
//     ],
//     enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
//     files: [
//       InlineFile.fromDataURL("data:text/plain;name=one.txt;base64,b25l=="),
//     ],
//     durations: [
//       Duration.fromISOString("PT2H3M4S"),
//       Duration.fromISOString("PT1H2M3S"),
//     ],
//   });

//   const thing = await models.thing.update(
//     { id: created.id },
//     {
//       texts: ["Keel", "Weave"],
//       numbers: [1, 2, 3],
//       booleans: [true, true, false],
//       dates: [
//         new Date(2001, 1, 2, 0, 0, 0, 0),
//         new Date(2002, 31, 12, 12, 45, 0, 0),
//         new Date(2003, 13, 3, 0, 0, 0, 0),
//       ],
//       timestamps: [
//         new Date("2023-01-02 23:00:30"),
//         new Date("2023-11-13 06:17:30.123"),
//         new Date("2024-02-01 23:00:30"),
//       ],
//       enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
//       files: [
//         InlineFile.fromDataURL("data:text/plain;name=two.txt;base64,dHdv=="),
//         InlineFile.fromDataURL(
//           "data:text/plain;name=three.txt;base64,dGhyZWU="
//         ),
//       ],
//       durations: [Duration.fromISOString("PT3H3M4S")],
//     }
//   );

//   expect(thing.texts).toHaveLength(2);
//   expect(thing.texts![0]).toEqual("Keel");
//   expect(thing.texts![1]).toEqual("Weave");

//   expect(thing.numbers).toHaveLength(3);
//   expect(thing.numbers![0]).toEqual(1);
//   expect(thing.numbers![1]).toEqual(2);
//   expect(thing.numbers![2]).toEqual(3);

//   expect(thing.booleans).toHaveLength(3);
//   expect(thing.booleans![0]).toEqual(true);
//   expect(thing.booleans![1]).toEqual(true);
//   expect(thing.booleans![2]).toEqual(false);

//   expect(thing.dates).toHaveLength(3);
//   expect(thing.dates![0]).toEqual(new Date(2001, 1, 2, 0, 0, 0, 0));
//   expect(thing.dates![1]).toEqual(new Date(2002, 31, 12, 0, 0, 0, 0));
//   expect(thing.dates![2]).toEqual(new Date(2003, 13, 3, 0, 0, 0, 0));

//   expect(thing.timestamps).toHaveLength(3);
//   expect(thing.timestamps![0]).toEqual(new Date("2023-01-02 23:00:30"));
//   expect(thing.timestamps![1]).toEqual(new Date("2023-11-13 06:17:30.123"));
//   expect(thing.timestamps![2]).toEqual(new Date("2024-02-01 23:00:30.000"));

//   expect(thing.enums).toHaveLength(3);
//   expect(thing.enums![0]).toEqual(MyEnum.One);
//   expect(thing.enums![1]).toEqual(MyEnum.Two);
//   expect(thing.enums![2]).toEqual(MyEnum.Three);

//   expect(thing.files).toHaveLength(2);

//   expect(thing.files![0].contentType).toEqual("text/plain");
//   expect(thing.files![0].filename).toEqual("two.txt");
//   expect(thing.files![0].size).toEqual(3);
//   const contents1 = await thing.files![0].read();
//   expect(contents1?.toString("utf-8")).toEqual("two");

//   expect(thing.files![1].contentType).toEqual("text/plain");
//   expect(thing.files![1].filename).toEqual("three.txt");
//   expect(thing.files![1].size).toEqual(5);
//   const contents2 = await thing.files![1].read();
//   expect(contents2?.toString("utf-8")).toEqual("three");

//   expect(thing.durations).toHaveLength(1);
//   expect(thing.durations![0]).toEqual({
//     hours: 3,
//     minutes: 3,
//     seconds: 4,
//   });
// });

// test("array functions - modelapi - text query", async () => {
//   const t1 = await models.thing.create({
//     texts: ["Keel", "Weave"],
//   });

//   const t2 = await models.thing.create({
//     texts: ["Keel", "Weave", "Keelson", "Keeler"],
//   });

//   const t3 = await models.thing.create({
//     texts: ["Keel", "Weave"],
//   });

//   const t4 = await models.thing.create({
//     texts: null,
//   });

//   const t5 = await models.thing.create({
//     texts: [],
//   });

//   const t6 = await models.thing.create({
//     texts: ["Weave", "Keel"],
//   });

//   const t7 = await models.thing.create({
//     texts: ["Keelson", "Keelson"],
//   });

//   const things1 = await models.thing.findMany({
//     where: {
//       texts: {
//         equals: ["Keel", "Weave"],
//       },
//     },
//   });

//   expect(things1).toHaveLength(2);
//   expect(things1).toEqual(expect.arrayContaining([t1, t3]));

//   const things2 = await models.thing.findMany({
//     where: {
//       texts: {
//         notEquals: ["Keel", "Weave"],
//       },
//     },
//   });

//   expect(things2).toHaveLength(5);
//   expect(things2).toEqual(expect.arrayContaining([t2, t4, t5, t6, t7]));

//   const things3 = await models.thing.findMany({
//     where: {
//       texts: {
//         equals: null,
//       },
//     },
//   });

//   expect(things3).toHaveLength(1);
//   expect(things3[0].id).toEqual(t4.id);

//   const things4 = await await models.thing.findMany({
//     where: {
//       texts: {
//         notEquals: null,
//       },
//     },
//   });

//   expect(things4).toHaveLength(6);
//   expect(things4).toEqual(expect.arrayContaining([t1, t2, t3, t5, t6, t7]));

//   const things5 = await models.thing.findMany({
//     where: {
//       texts: {
//         equals: [],
//       },
//     },
//   });

//   expect(things5).toHaveLength(1);
//   expect(things5[0].id).toEqual(t5.id);

//   const ads = await models.thing.findMany({
//     where: { texts: { any: { equals: "Weave" } } },
//   });

//   const things6 = await models.thing.findMany({
//     where: {
//       texts: {
//         notEquals: [],
//       },
//     },
//   });

//   expect(things6).toHaveLength(6);
//   expect(things6).toEqual(expect.arrayContaining([t1, t2, t3, t4, t6, t7]));

//   const things7 = await models.thing.findMany({
//     where: {
//       texts: {
//         any: {
//           equals: "Weave",
//         },
//       },
//     },
//   });

//   expect(things7).toHaveLength(4);
//   expect(things7).toEqual(expect.arrayContaining([t1, t2, t3, t6]));

//   const things8 = await models.thing.findMany({
//     where: {
//       texts: {
//         all: {
//           equals: "Keelson",
//         },
//       },
//     },
//   });

//   expect(things8).toHaveLength(2);
//   expect(things8).toEqual(expect.arrayContaining([t5, t7]));

//   const things9 = await models.thing.findMany({
//     where: {
//       texts: {
//         any: {
//           equals: "Keelson",
//           notEquals: "Weave",
//         },
//       },
//     },
//   });

//   expect(things9).toHaveLength(1);
//   expect(things9[0].id).toEqual(t7.id);

//   const things10 = await models.thing.findMany({
//     where: {
//       texts: {
//         any: {
//           notEquals: "Weave",
//         },
//       },
//     },
//   });

//   expect(things10).toHaveLength(2);
//   expect(things10).toEqual(expect.arrayContaining([t5, t7]));

//   const things11 = await models.thing.findMany({
//     where: {
//       texts: {
//         all: {
//           notEquals: "Keelson",
//         },
//       },
//     },
//   });

//   expect(things11).toHaveLength(4);
//   expect(things11).toEqual(expect.arrayContaining([t1, t2, t3, t6]));
// });

// test("array functions - list action implicit querying - number", async () => {
//   const t1 = await models.thing.create({
//     numbers: [1, 2],
//   });

//   const t2 = await models.thing.create({
//     numbers: [1, 2, 3, 4],
//   });

//   const t3 = await models.thing.create({
//     numbers: [1, 2],
//   });

//   const t4 = await models.thing.create({
//     numbers: null,
//   });

//   const t5 = await models.thing.create({
//     numbers: [],
//   });

//   const t6 = await models.thing.create({
//     numbers: [2, 1],
//   });

//   const things = await models.thing.findMany({
//     where: {
//       numbers: {
//         equals: [1, 2],
//       },
//     },
//   });

//   expect(things).toHaveLength(2);
//   expect(things).toEqual(expect.arrayContaining([t1, t3]));
// });

// test("array functions - list action implicit querying - date", async () => {
//   const t1 = await models.thing.create({
//     dates: [new Date(2024, 1, 1, 0, 0, 0, 0), new Date(2024, 1, 2, 0, 0, 0, 0)],
//   });

//   const t2 = await models.thing.create({
//     dates: [
//       new Date(2024, 1, 1, 0, 0, 0, 0),
//       new Date(2024, 1, 2, 0, 0, 0, 0),
//       new Date(2024, 1, 3, 0, 0, 0, 0),
//     ],
//   });

//   const t3 = await models.thing.create({
//     dates: [new Date(2024, 1, 1, 0, 0, 0, 0), new Date(2024, 1, 2, 0, 0, 0, 0)],
//   });

//   const t4 = await models.thing.create({
//     dates: null,
//   });

//   const t5 = await models.thing.create({
//     dates: [],
//   });

//   const t6 = await models.thing.create({
//     dates: [new Date(2024, 1, 2, 0, 0, 0, 0), new Date(2024, 1, 1, 0, 0, 0, 0)],
//   });

//   const things = await models.thing.findMany({
//     where: {
//       dates: {
//         equals: [
//           new Date(2024, 1, 1, 0, 0, 0, 0),
//           new Date(2024, 1, 2, 0, 0, 0, 0),
//         ],
//       },
//     },
//   });

//   expect(things).toHaveLength(2);
//   expect(things).toEqual(expect.arrayContaining([t1, t3]));
// });

// test("array functions - list action implicit querying - timestamp", async () => {
//   const t1 = await models.thing.create({
//     timestamps: [
//       new Date(2024, 1, 1, 30, 45, 50, 0),
//       new Date(2024, 1, 2, 59, 0, 0, 0),
//     ],
//   });

//   const t2 = await models.thing.create({
//     timestamps: [
//       new Date(2024, 1, 1, 30, 45, 50, 0),
//       new Date(2024, 1, 2, 59, 0, 0, 0),
//       new Date(2024, 1, 3, 0, 0, 0, 0),
//     ],
//   });

//   const t3 = await models.thing.create({
//     timestamps: [
//       new Date(2024, 1, 1, 30, 45, 50, 0),
//       new Date(2024, 1, 2, 59, 0, 0, 0),
//     ],
//   });

//   const t4 = await models.thing.create({
//     timestamps: null,
//   });

//   const t5 = await models.thing.create({
//     timestamps: [],
//   });

//   const t6 = await models.thing.create({
//     timestamps: [
//       new Date(2024, 1, 2, 59, 0, 0, 0),
//       new Date(2024, 1, 1, 30, 45, 50, 0),
//     ],
//   });

//   const things = await models.thing.findMany({
//     where: {
//       timestamps: {
//         equals: [
//           new Date(2024, 1, 1, 30, 45, 50, 0),
//           new Date(2024, 1, 2, 59, 0, 0, 0),
//         ],
//       },
//     },
//   });

//   expect(things).toHaveLength(2);
//   expect(things).toEqual(expect.arrayContaining([t1, t3]));
// });

// test("array functions - list action implicit querying - enums", async () => {
//   const t1 = await models.thing.create({
//     enums: [MyEnum.One, MyEnum.Two],
//   });

//   const t2 = await models.thing.create({
//     enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
//   });

//   const t3 = await models.thing.create({
//     enums: [MyEnum.One, MyEnum.Two],
//   });

//   const t4 = await models.thing.create({
//     enums: null,
//   });

//   const t5 = await models.thing.create({
//     enums: [],
//   });

//   const t6 = await models.thing.create({
//     enums: [MyEnum.Two, MyEnum.One],
//   });

//   const things = await models.thing.findMany({
//     where: {
//       enums: {
//         equals: [MyEnum.One, MyEnum.Two],
//       },
//     },
//   });

//   expect(things).toHaveLength(2);
//   expect(things).toEqual(expect.arrayContaining([t1, t3]));
// });

test("array functions - create with empty hooks", async () => {
  const thing = await actions.createThingEmpty({
    texts: ["Keel", "Weave"],
    numbers: [1, 2, 3],
    booleans: [true, true, false],
    timestamps: [
      new Date("2023-01-02 23:00:30"),
      new Date("2023-11-13 06:17:30.123"),
      new Date("2024-02-01 23:00:30"),
    ],
    enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
    files: [
      InlineFile.fromDataURL("data:text/plain;name=one.txt;base64,b25l=="),
      InlineFile.fromDataURL("data:text/plain;name=two.txt;base64,dHdv=="),
    ],
    durations: [
      Duration.fromISOString("PT2H3M4S"),
      Duration.fromISOString("PT1H2M3S"),
    ],
  });

  expect(thing?.texts).toEqual(["Keel", "Weave"]);
  expect(thing?.numbers).toEqual([1, 2, 3]);
  expect(thing?.booleans).toEqual([true, true, false]);
  expect(thing?.timestamps).toEqual([
    new Date("2023-01-02 23:00:30"),
    new Date("2023-11-13 06:17:30.123"),
    new Date("2024-02-01 23:00:30"),
  ]);
  expect(thing?.enums).toEqual([MyEnum.One, MyEnum.Two, MyEnum.Three]);

  expect(thing?.files).toHaveLength(2);

  expect(thing?.files?.[0].contentType).toEqual("text/plain");
  expect(thing?.files?.[0].filename).toEqual("one.txt");
  expect(thing?.files?.[0].size).toEqual(3);
  const contents1 = await thing?.files?.[0].read();
  expect(contents1?.toString("utf-8")).toEqual("one");

  expect(thing?.files?.[1].contentType).toEqual("text/plain");
  expect(thing?.files?.[1].filename).toEqual("two.txt");
  expect(thing?.files?.[1].size).toEqual(3);
  const contents2 = await thing?.files?.[1].read();
  expect(contents2?.toString("utf-8")).toEqual("two");

  expect(thing?.durations).toEqual([
    {
      hours: 2,
      minutes: 3,
      seconds: 4,
    },
    {
      hours: 1,
      minutes: 2,
      seconds: 3,
    },
  ]);
});

// test("array fields - update with empty hooks", async () => {
//   const thing = await actions.createThingEmpty({
//     texts: ["Keel", "Weave"],
//     numbers: [1, 2, 3],
//     booleans: [true, true, false],
//     timestamps: [
//       new Date("2023-01-02 23:00:30"),
//       new Date("2023-11-13 06:17:30.123"),
//       new Date("2024-02-01 23:00:30"),
//     ],
//     enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
//     files: [
//       InlineFile.fromDataURL("data:text/plain;name=one.txt;base64,b25l=="),
//       InlineFile.fromDataURL("data:text/plain;name=two.txt;base64,dHdv=="),
//     ],
//     durations: [
//       Duration.fromISOString("PT2H3M4S"),
//       Duration.fromISOString("PT1H2M3S"),
//     ],
//   });

//   expect(thing?.texts).toEqual(["Keel", "Weave"]);
//   expect(thing?.numbers).toEqual([1, 2, 3]);
//   expect(thing?.booleans).toEqual([true, true, false]);
//   expect(thing?.timestamps).toEqual([
//     new Date("2023-01-02 23:00:30"),
//     new Date("2023-11-13 06:17:30.123"),
//     new Date("2024-02-01 23:00:30"),
//   ]);
//   expect(thing?.enums).toEqual([MyEnum.One, MyEnum.Two, MyEnum.Three]);

//   expect(thing?.files).toHaveLength(2);

//   expect(thing?.files?.[0].contentType).toEqual("text/plain");
//   expect(thing?.files?.[0].filename).toEqual("one.txt");
//   expect(thing?.files?.[0].size).toEqual(3);
//   const contents1 = await thing?.files?.[0].read();
//   expect(contents1?.toString("utf-8")).toEqual("one");

//   expect(thing?.files?.[1].contentType).toEqual("text/plain");
//   expect(thing?.files?.[1].filename).toEqual("two.txt");
//   expect(thing?.files?.[1].size).toEqual(3);
//   const contents2 = await thing?.files?.[1].read();
//   expect(contents2?.toString("utf-8")).toEqual("two");

//   expect(thing?.durations).toEqual([
//     {
//       hours: 2,
//       minutes: 3,
//       seconds: 4,
//     },
//     {
//       hours: 1,
//       minutes: 2,
//       seconds: 3,
//     },
//   ]);
// });

// test("array functions - create with changes in hooks", async () => {
//   const thing = await actions.createThing({
//     texts: ["Keel", "Weave"],
//     numbers: [1, 2, 3],
//     booleans: [true, true, false],
//     timestamps: [
//       new Date("2023-01-02 23:00:30"),
//       new Date("2023-11-13 06:17:30.123"),
//       new Date("2024-02-01 23:00:30"),
//     ],
//     enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
//     files: [
//       InlineFile.fromDataURL("data:text/plain;name=one.txt;base64,b25l=="),
//       InlineFile.fromDataURL("data:text/plain;name=two.txt;base64,dHdv=="),
//     ],
//     durations: [
//       Duration.fromISOString("PT2H3M4S"),
//       Duration.fromISOString("PT1H2M3S"),
//     ],
//   });

//   expect(thing?.texts).toEqual(["Keel"]);
//   expect(thing?.numbers).toEqual([1, 2]);
//   expect(thing?.booleans).toEqual([true, true]);
//   expect(thing?.timestamps).toEqual([
//     new Date("2023-01-02 23:00:30"),
//     new Date("2023-11-13 06:17:30.123"),
//   ]);
//   expect(thing?.enums).toEqual([MyEnum.One, MyEnum.Two]);

//   expect(thing?.files).toHaveLength(1);

//   expect(thing?.files?.[0].contentType).toEqual("text/plain");
//   expect(thing?.files?.[0].filename).toEqual("one.txt");
//   expect(thing?.files?.[0].size).toEqual(3);
//   const contents1 = await thing?.files?.[0].read();
//   expect(contents1?.toString("utf-8")).toEqual("one");

//   expect(thing?.durations).toEqual([
//     {
//       hours: 2,
//       minutes: 3,
//       seconds: 4,
//     },
//   ]);
// });

// test("array functions - custom write function", async () => {
//   const msg = await actions.writeThing({
//     texts: ["Keel", "Weave"],
//     numbers: [1, 2, 3],
//     booleans: [true, true, false],
//     timestamps: [
//       new Date("2023-01-02 23:00:30"),
//       new Date("2023-11-13 06:17:30.123"),
//       new Date("2024-02-01 23:00:30"),
//     ],
//     enums: [MyEnum.One, MyEnum.Two, MyEnum.Three],
//     files: [
//       InlineFile.fromDataURL("data:text/plain;name=one.txt;base64,b25l=="),
//       InlineFile.fromDataURL("data:text/plain;name=two.txt;base64,dHdv=="),
//     ],
//     durations: [
//       Duration.fromISOString("PT2H3M4S"),
//       Duration.fromISOString("PT1H2M3S"),
//     ],
//   });

//   expect(msg.thing.texts).toEqual(["Keel", "Weave"]);
//   expect(msg.thing.numbers).toEqual([1, 2, 3]);
//   expect(msg.thing.booleans).toEqual([true, true, false]);
//   expect(msg.thing.timestamps).toEqual([
//     new Date("2023-01-02 23:00:30"),
//     new Date("2023-11-13 06:17:30.123"),
//     new Date("2024-02-01 23:00:30"),
//   ]);
//   expect(msg.thing.enums).toEqual([MyEnum.One, MyEnum.Two, MyEnum.Three]);

//   expect(msg.thing.files).toHaveLength(2);

//   expect(msg.thing.files?.[0].contentType).toEqual("text/plain");
//   expect(msg.thing.files?.[0].filename).toEqual("one.txt");
//   expect(msg.thing.files?.[0].size).toEqual(3);
//   const contents1 = await msg.thing.files?.[0].read();
//   expect(contents1?.toString("utf-8")).toEqual("one");

//   expect(msg.thing.files?.[1].contentType).toEqual("text/plain");
//   expect(msg.thing.files?.[1].filename).toEqual("two.txt");
//   expect(msg.thing.files?.[1].size).toEqual(3);
//   const contents2 = await msg.thing.files?.[1].read();
//   expect(contents2?.toString("utf-8")).toEqual("two");

//   expect(msg.thing.durations).toEqual([
//     {
//       hours: 2,
//       minutes: 3,
//       seconds: 4,
//     },
//     {
//       hours: 1,
//       minutes: 2,
//       seconds: 3,
//     },
//   ]);

//   const getThing = await actions.getThing({ id: msg.thing.id });

//   expect(getThing?.texts).toEqual(["Keel", "Weave"]);
//   expect(getThing?.numbers).toEqual([1, 2, 3]);
//   expect(getThing?.booleans).toEqual([true, true, false]);
//   expect(getThing?.timestamps).toEqual([
//     new Date("2023-01-02 23:00:30"),
//     new Date("2023-11-13 06:17:30.123"),
//     new Date("2024-02-01 23:00:30"),
//   ]);
//   expect(getThing?.enums).toEqual([MyEnum.One, MyEnum.Two, MyEnum.Three]);

//   expect(getThing?.files).toHaveLength(2);

//   expect(getThing?.files?.[0].contentType).toEqual("text/plain");
//   expect(getThing?.files?.[0].filename).toEqual("one.txt");
//   expect(getThing?.files?.[0].size).toEqual(3);
//   const contents3 = await getThing?.files?.[0].read();
//   expect(contents3?.toString("utf-8")).toEqual("one");

//   expect(getThing?.files?.[1].contentType).toEqual("text/plain");
//   expect(getThing?.files?.[1].filename).toEqual("two.txt");
//   expect(getThing?.files?.[1].size).toEqual(3);
//   const contents4 = await getThing?.files?.[1].read();
//   expect(contents4?.toString("utf-8")).toEqual("two");

//   expect(getThing?.durations).toEqual([
//     {
//       hours: 2,
//       minutes: 3,
//       seconds: 4,
//     },
//     {
//       hours: 1,
//       minutes: 2,
//       seconds: 3,
//     },
//   ]);
// });

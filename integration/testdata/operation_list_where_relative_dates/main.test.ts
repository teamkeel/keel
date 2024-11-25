import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("List Where filters - date - filters correctly", async () => {
  const today = new Date();
  const tomorrow = new Date();
  tomorrow.setDate(today.getDate() + 1);
  const yesterday = new Date();
  yesterday.setDate(today.getDate() - 1);
  const aWeekAgo = new Date();
  aWeekAgo.setDate(today.getDate() - 7);

  await actions.createPost({ title: "Today", aDate: today });
  await actions.createPost({ title: "Tomorrow", aDate: tomorrow });
  await actions.createPost({ title: "Yesterday", aDate: yesterday });
  await actions.createPost({ title: "A Week Ago", aDate: aWeekAgo });

  const r1 = await actions.listPostsByDate({
    where: {
      aDate: {
        beforeRelative: "tomorrow",
      },
    },
  });

  const posts1 = r1!.results.map((x) => x.title);
  expect(posts1).toHaveLength(3);
  expect(posts1).toContain("Today");
  expect(posts1).toContain("Yesterday");
  expect(posts1).toContain("A Week Ago");

  const r2 = await actions.listPostsByDate({
    where: {
      aDate: {
        afterRelative: "today",
      },
    },
  });

  expect(r2.results.length).toEqual(1);
  expect(r2.results[0].title).toEqual("Tomorrow");

  const r3 = await actions.listPostsByDate({
    where: {
      aDate: {
        equalsRelative: "last 7 complete days",
      },
    },
  });

  const posts3 = r3!.results.map((x) => x.title);
  expect(posts3).toHaveLength(2);
  expect(posts3).toContain("Yesterday");
  expect(posts3).toContain("A Week Ago");
});

test("List Where filters - timestamps - filters correctly", async () => {
  const now = new Date();
  const tomorrow = new Date();
  tomorrow.setDate(now.getDate() + 1);
  const yesterday = new Date();
  yesterday.setDate(now.getDate() - 1);

  await actions.createPost({ title: "Now", aTimestamp: now });
  await actions.createPost({ title: "Tomorrow", aTimestamp: tomorrow });
  await actions.createPost({ title: "Yesterday", aTimestamp: yesterday });

  const r1 = await actions.listPostsByTimestamp({
    where: {
      aTimestamp: {
        beforeRelative: "tomorrow",
      },
    },
  });

  const posts1 = r1!.results.map((x) => x.title);
  expect(posts1).toHaveLength(2);
  expect(posts1).toContain("Now");
  expect(posts1).toContain("Yesterday");

  const r2 = await actions.listPostsByTimestamp({
    where: {
      aTimestamp: {
        afterRelative: "now",
      },
    },
  });

  expect(r2.results.length).toEqual(1);
  expect(r2.results[0].title).toEqual("Tomorrow");

  const r3 = await actions.listPostsByTimestamp({
    where: {
      aTimestamp: {
        equalsRelative: "last week",
      },
    },
  });

  const posts3 = r3!.results.map((x) => x.title);
  expect(posts3).toHaveLength(2);
  expect(posts3).toContain("Now");
  expect(posts3).toContain("Yesterday");
});

test("List Where filters - date - with timezones", async () => {
  // all these dates go in UTC
  const today = new Date();
  const tomorrow = new Date();
  tomorrow.setDate(today.getDate() + 1);
  const yesterday = new Date();
  yesterday.setDate(today.getDate() - 1);

  await actions.createPost({ title: "Yesterday", aDate: yesterday });
  await actions.createPost({ title: "Today", aDate: today });
  await actions.createPost({ title: "Tomorrow", aDate: tomorrow });

  // Pacific/Kiritimati = UTC +14h
  // Pacific/Midway = UTC -11h

  // difference between Kiritimati and Midway = 25 hrs, therefore at any UTC time, Midway should have 1 less results as the
  // date would be 1 day behind Kiritimati.
  const r1 = await actions.withTimezone("Pacific/Kiritimati").listPostsByDate({
    where: {
      aDate: {
        beforeRelative: "today",
      },
    },
  });
  const r2 = await actions.withTimezone("Pacific/Midway").listPostsByDate({
    where: {
      aDate: {
        beforeRelative: "today",
      },
    },
  });
  expect(r2.results.length).toBeLessThan(r1.results.length);
});

test("List Where filters - date - with invalid timezone", async () => {
  await expect(
    actions.withTimezone("Europe/SuttonColdfield").listPostsByDate({
      where: {
        aDate: {
          beforeRelative: "today",
        },
      },
    })
  ).toHaveError({
    code: "ERR_INPUT_MALFORMED",
    message:
      "invalid Time-Zone header: unknown time zone Europe/SuttonColdfield",
  });
});

test("List Where filters - with hook", async () => {
  actions.withTimezone("UTC");
  // all these dates go in UTC
  const today = new Date();
  const tomorrow = new Date();
  tomorrow.setDate(today.getDate() + 1);
  const yesterday = new Date();
  yesterday.setDate(today.getDate() - 1);

  await actions.createPost({ title: "Yesterday", aDate: yesterday });
  await actions.createPost({ title: "Today", aDate: today });
  await actions.createPost({ title: "Tomorrow", aDate: tomorrow });

  // the hook sets beforeRelative = "today"
  const r1 = await actions.listPostsByDateWithHook({ where: {} });
  expect(r1.results.length).toEqual(1);
  expect(r1.results[0].title).toEqual("Yesterday");

  // the hook sets beforeRelative = "today"
  const r2 = await actions.listPostsByDateWithHook({
    where: {
      aDate: {
        beforeRelative: "tomorrow",
      },
    },
  });
  expect(r2.results.length).toEqual(1);
  expect(r2.results[0].title).toEqual("Yesterday");
});

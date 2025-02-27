import { test, expect, beforeEach, beforeAll } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeAll(async () => {
  const teamSA = await models.team.create({ name: "South Africa" });
  const teamUK = await models.team.create({ name: "United Kingdom" });
  const teamAus = await models.team.create({
    name: "Australia",
    disqualified: true,
  });
  await models.contestant.create({
    name: "Donald",
    gold: 2,
    silver: 4,
    bronze: 5,
    teamId: teamUK.id,
  });
  await models.contestant.create({
    name: "Bongani",
    gold: 4,
    silver: 5,
    bronze: 7,
    teamId: teamSA.id,
  });
  await models.contestant.create({
    name: "John",
    gold: 4,
    silver: 1,
    bronze: 3,
    teamId: teamUK.id,
  });
  await models.contestant.create({
    name: "Stoffel",
    gold: 4,
    silver: 5,
    bronze: 10,
    teamId: teamSA.id,
  });
  await models.contestant.create({
    name: "Mary",
    gold: 7,
    silver: 1,
    bronze: 3,
    teamId: teamUK.id,
  });
  await models.contestant.create({
    name: "Johannes",
    disqualified: true,
    gold: 3,
    silver: 1,
    bronze: 3,
    teamId: teamSA.id,
  });
  await models.contestant.create({
    name: "Addison",
    gold: 6,
    silver: 6,
    bronze: 6,
    teamId: teamAus.id,
  });
});

test("orderby", async () => {
  const winners = await actions.listRankings();

  expect(winners.pageInfo.count).toEqual(5);
  expect(winners.pageInfo.totalCount).toEqual(5);
  expect(winners.pageInfo.hasNextPage).toEqual(false);
  expect(winners.results[0].name).toEqual("Mary");
  expect(winners.results[1].name).toEqual("Stoffel");
  expect(winners.results[2].name).toEqual("Bongani");
  expect(winners.results[3].name).toEqual("John");
  expect(winners.results[4].name).toEqual("Donald");
});

test("orderby - get first - top 3 winners", async () => {
  const winners = await actions.listRankings({
    first: 3,
  });

  expect(winners.pageInfo.count).toEqual(3);
  expect(winners.pageInfo.totalCount).toEqual(5);
  expect(winners.pageInfo.hasNextPage).toEqual(true);
  expect(winners.results[0].name).toEqual("Mary");
  expect(winners.results[1].name).toEqual("Stoffel");
  expect(winners.results[2].name).toEqual("Bongani");
});

test("orderby - get top 3 - with offset pagination", async () => {
  const winners = await actions.listRankings({
    limit: 3,
  });

  expect(winners.pageInfo.count).toEqual(3);
  expect(winners.pageInfo.pageNumber).toEqual(1);
  expect(winners.pageInfo.totalCount).toEqual(5);
  expect(winners.pageInfo.hasNextPage).toEqual(true);
  expect(winners.results[0].name).toEqual("Mary");
  expect(winners.results[1].name).toEqual("Stoffel");
  expect(winners.results[2].name).toEqual("Bongani");
});

test("orderby - get second - with offset pagination", async () => {
  const winners = await actions.listRankings({
    limit: 1,
    offset: 1,
  });

  expect(winners.pageInfo.count).toEqual(1);
  expect(winners.pageInfo.pageNumber).toEqual(2);
  expect(winners.pageInfo.totalCount).toEqual(5);
  expect(winners.pageInfo.hasNextPage).toEqual(true);
  expect(winners.results[0].name).toEqual("Stoffel");
});

test("orderby - with filter - saffer rankings", async () => {
  const saffers = await actions.listRankings({
    where: {
      team: { name: { equals: "South Africa" } },
    },
  });

  expect(saffers.pageInfo.count).toEqual(2);
  expect(saffers.pageInfo.totalCount).toEqual(2);
  expect(saffers.pageInfo.hasNextPage).toEqual(false);
  expect(saffers.results[0].name).toEqual("Stoffel");
  expect(saffers.results[1].name).toEqual("Bongani");
});

test("orderby - get first with filter - saffer winner", async () => {
  const saffers = await actions.listRankings({
    first: 1,
    where: {
      team: { name: { equals: "South Africa" } },
    },
  });

  expect(saffers.pageInfo.count).toEqual(1);
  expect(saffers.pageInfo.totalCount).toEqual(2);
  expect(saffers.pageInfo.hasNextPage).toEqual(true);
  expect(saffers.results[0].name).toEqual("Stoffel");
});

test("orderby - get first and after cursor - fourth place", async () => {
  const winners = await actions.listRankings({
    first: 3,
  });

  const cursor = winners.pageInfo.endCursor;
  const fourth = await actions.listRankings({
    first: 1,
    after: cursor,
  });

  expect(fourth.pageInfo.count).toEqual(1);
  expect(fourth.pageInfo.totalCount).toEqual(5);
  expect(fourth.pageInfo.hasNextPage).toEqual(true);
  expect(fourth.results[0].name).toEqual("John");
});

test("orderby - before cursor - better than John", async () => {
  const winners = await actions.listRankings();

  const cursor = winners.results[3].id;
  const betterThanJohn = await actions.listRankings({
    before: cursor,
    last: 3,
  });

  expect(betterThanJohn.pageInfo.count).toEqual(3);
  expect(betterThanJohn.pageInfo.totalCount).toEqual(5);
  expect(betterThanJohn.pageInfo.hasNextPage).toEqual(false);
  expect(betterThanJohn.results[0].name).toEqual("Mary");
  expect(betterThanJohn.results[1].name).toEqual("Stoffel");
  expect(betterThanJohn.results[2].name).toEqual("Bongani");
});

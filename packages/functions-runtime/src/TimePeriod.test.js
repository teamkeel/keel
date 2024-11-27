import { test, expect } from "vitest";
const { TimePeriod } = require("./TimePeriod");

test("shorthands test", async () => {
  const today = TimePeriod.fromExpression("today");
  expect(today).toEqual({
    period: "day",
    value: 1,
    complete: true,
    offset: 0,
  });

  const tomorrow = TimePeriod.fromExpression("tomorrow");
  expect(tomorrow).toEqual({
    period: "day",
    value: 1,
    complete: true,
    offset: 1,
  });

  const yesterday = TimePeriod.fromExpression("yesterday");
  expect(yesterday).toEqual({
    period: "day",
    value: 1,
    complete: true,
    offset: -1,
  });

  const now = TimePeriod.fromExpression("now");
  expect(now).toEqual({
    period: "",
    value: 0,
    complete: false,
    offset: 0,
  });
  expect(now.periodStartSQL()).toEqual("(NOW())");
  expect(now.periodEndSQL()).toEqual("(NOW())");
});

test("next test", async () => {
  let period = TimePeriod.fromExpression("next day");
  expect(period).toEqual({
    period: "day",
    value: 1,
    complete: false,
    offset: 0,
  });

  period = TimePeriod.fromExpression("next complete day");
  expect(period).toEqual({
    period: "day",
    value: 1,
    complete: true,
    offset: 1,
  });

  period = TimePeriod.fromExpression("next 5 complete day");
  expect(period).toEqual({
    period: "day",
    value: 5,
    complete: true,
    offset: 1,
  });

  period = TimePeriod.fromExpression("next 5 months");
  expect(period).toEqual({
    period: "month",
    value: 5,
    complete: false,
    offset: 0,
  });

  period = TimePeriod.fromExpression("next 2 complete years");
  expect(period).toEqual({
    period: "year",
    value: 2,
    complete: true,
    offset: 1,
  });
});

test("last test", async () => {
  let period = TimePeriod.fromExpression("last day");
  expect(period).toEqual({
    period: "day",
    value: 1,
    complete: false,
    offset: -1,
  });

  period = TimePeriod.fromExpression("last complete day");
  expect(period).toEqual({
    period: "day",
    value: 1,
    complete: true,
    offset: -1,
  });

  period = TimePeriod.fromExpression("last 5 complete day");
  expect(period).toEqual({
    period: "day",
    value: 5,
    complete: true,
    offset: -5,
  });

  period = TimePeriod.fromExpression("last 5 months");
  expect(period).toEqual({
    period: "month",
    value: 5,
    complete: false,
    offset: -5,
  });

  period = TimePeriod.fromExpression("last 2 complete years");
  expect(period).toEqual({
    period: "year",
    value: 2,
    complete: true,
    offset: -2,
  });

  period = TimePeriod.fromExpression("last complete year");
  expect(period).toEqual({
    period: "year",
    value: 1,
    complete: true,
    offset: -1,
  });
});

test("errors test", async () => {
  expect(() => TimePeriod.fromExpression("last test day")).toThrowError(
    "Invalid time period expression"
  );
  expect(() => TimePeriod.fromExpression("today now")).toThrowError(
    "Invalid time period expression"
  );
  expect(() => TimePeriod.fromExpression("last -5 days")).toThrowError(
    "Invalid time period expression"
  );
  expect(() => TimePeriod.fromExpression("5 mont")).toThrowError(
    "Invalid time period expression"
  );
  expect(() => TimePeriod.fromExpression("5 days")).toThrowError(
    "Time period expression must start with this, next, or last"
  );
});

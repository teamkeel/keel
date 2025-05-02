import { test, expect } from "vitest";
import { Duration } from "./Duration";

test("fromISOString test", async () => {
  const fullDate = Duration.fromISOString("P1Y2M3DT4H5M6S");
  expect(fullDate.toISOString()).toEqual("P1Y2M3DT4H5M6S");
  expect(fullDate.toPostgres()).toEqual(
    "1 year 2 months 3 days 4 hours 5 minutes 6 seconds"
  );
  const dateOnly = Duration.fromISOString("P2Y3M4D");
  expect(dateOnly.toISOString()).toEqual("P2Y3M4D");
  expect(dateOnly.toPostgres()).toEqual("2 years 3 months 4 days");
  const timeOnly = Duration.fromISOString("PT4H5M6S");
  expect(timeOnly.toISOString()).toEqual("PT4H5M6S");
  expect(timeOnly.toPostgres()).toEqual("4 hours 5 minutes 6 seconds");
  const years = Duration.fromISOString("P10Y");
  expect(years.toISOString()).toEqual("P10Y");
  expect(years.toPostgres()).toEqual("10 years");
  const months = Duration.fromISOString("P20M");
  expect(months.toISOString()).toEqual("P20M");
  expect(months.toPostgres()).toEqual("20 months");
  const days = Duration.fromISOString("P31D");
  expect(days.toISOString()).toEqual("P31D");
  expect(days.toPostgres()).toEqual("31 days");
  const hours = Duration.fromISOString("PT4H");
  expect(hours.toISOString()).toEqual("PT4H");
  expect(hours.toPostgres()).toEqual("4 hours");
  const minutes = Duration.fromISOString("PT61M");
  expect(minutes.toISOString()).toEqual("PT61M");
  expect(minutes.toPostgres()).toEqual("61 minutes");
  const seconds = Duration.fromISOString("PT76S");
  expect(seconds.toISOString()).toEqual("PT76S");
  expect(seconds.toPostgres()).toEqual("76 seconds");
});

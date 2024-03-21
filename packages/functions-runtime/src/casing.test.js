import { test, expect } from "vitest";
import { camelCaseObject, snakeCaseObject } from "./casing";

const cases = {
  FROM_SNAKE: {
    input: {
      id: "123",
      slack_id: "xxx_2929",
      api_key: "1234",
      test_11: "1234",
      test_1_test: "1234",
      test12: "1234",
    },
    expected: {
      id: "123",
      slackId: "xxx_2929",
      apiKey: "1234",
      test11: "1234",
      test1Test: "1234",
      test12: "1234",
    },
  },
  FROM_CAMEL: {
    input: {
      id: "123",
      slackId: "xxx_2929",
      apiKey: "1234",
      test1: "test",
      testA1: "test",
      test1Test: "test",
      test20: "test",
    },
    expected: {
      id: "123",
      slack_id: "xxx_2929",
      api_key: "1234",
      test_1: "test",
      test_a_1: "test",
      test_1_test: "test",
      test_20: "test",
    },
  },
};

Object.entries(cases).map(([key, { input, expected }]) => {
  test(key, () => {
    let result;
    if (key === "FROM_SNAKE") {
      result = camelCaseObject(input);
    } else {
      result = snakeCaseObject(input);
    }

    expect(result).toEqual(expected);
  });
});

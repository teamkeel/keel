import { test, expect } from "vitest";
import { camelCaseObject } from "./casing";

const cases = {
  FROM_SNAKE: {
    input: {
      id: "123",
      slack_id: "xxx_2929",
      api_key: "1234"
    },
    expected: {
      id: "123",
      slackId: "xxx_2929",
      apiKey: "1234"
    },
  },
}

Object.entries(cases).map(([key, { input, expected }]) => {
  test(key, () => {  
    const result = camelCaseObject(input);
  
    expect(result).toEqual(expected);
  });
})

import { describe, expect, test } from "vitest";
import { grid, GridOptions } from "./grid";

// Use the usage input and return the ui config response
const testGridAPI = async <T>(options: GridOptions<T>) => {
  return (await grid(options as GridOptions<unknown>)).uiConfig;
};

describe("grid element", () => {
  describe("ui config", () => {
    test("simple render function", async () => {
      const res = await testGridAPI({
        data: [
          {
            name: "John",
          },
        ],
        render: (data) => ({
          title: data.name,
          extraField: "this should not be in the response",
        }),
      });

      expect(res.data).toEqual([
        {
          title: "John",
        },
      ]);
    });
  });
});

import { describe, expect, test } from "vitest";
import { grid, GridOptions, UiElementGrid } from "./grid";

// Use the usage input and return the ui config response
const testGridAPI = <T>(options: GridOptions<T>) => {
  return grid(options as GridOptions<unknown>).uiConfig;
};

describe("grid element", () => {
  describe("ui config", () => {
    test("simple render function", () => {
      const res = testGridAPI({
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

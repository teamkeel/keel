import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("select one element", () => {
  test("response data is typed correctly", () => {
    testFlow({}, async (ctx) => {
      const res = await ctx.ui.page("page", {
        content: [
          ctx.ui.select.one("name", {
            options: ["one", "two", "three"],
          }),
        ],
      });

      expectTypeOf(res.name).toBeString();
      expectTypeOf(res.name).toEqualTypeOf<"one" | "two" | "three">();
    });
  });
});

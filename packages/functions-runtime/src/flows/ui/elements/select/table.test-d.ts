import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("select table element", () => {
  test("multi select types", () => {
    testFlow({}, async (ctx) => {
      const thing: string = "foo";

      const res = await ctx.ui.page("page", {
        content: [
          ctx.ui.select.table("name", {
            data: [
              {
                thing: thing,
              },
              {
                thing: thing,
              },
            ],
          }),
        ],
      });

      expectTypeOf(res.name).toBeArray();
      expectTypeOf(res.name).branded.toEqualTypeOf<
        {
          thing: string;
        }[]
      >;
    });
  });

  test("single select types", () => {
    testFlow({}, async (ctx) => {
      const thing: number = 123;

      const res = await ctx.ui.page("page", {
        content: [
          ctx.ui.select.table("name", {
            data: [
              {
                thing: thing,
              },
              {
                thing: thing,
              },
            ],
            mode: "single",
          }),
        ],
      });

      expectTypeOf(res.name).toBeObject();
      expectTypeOf(res.name).branded.toEqualTypeOf<{
        thing: number;
      }>;
    });
  });
});

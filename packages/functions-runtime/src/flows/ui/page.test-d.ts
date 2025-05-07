import { describe, expectTypeOf, test } from "vitest";
import { _testFlow } from "../testingUtils";

describe("page element", () => {
  test("return data is typed correctly", () => {
    _testFlow({}, async (ctx) => {
      const res = await ctx.ui.page("page", {
        content: [
          ctx.ui.display.header(),
          ctx.ui.inputs.text("name"),
          ctx.ui.inputs.text("email"),
          ctx.ui.inputs.text("phone"),
          ctx.ui.display.divider(),
          ctx.ui.display.markdown(),
          ctx.ui.inputs.boolean("terms"),
        ],
      });

      expectTypeOf(res.name).toBeString();
      expectTypeOf(res.email).toBeString();
      expectTypeOf(res.phone).toBeString();
      expectTypeOf(res.terms).toBeBoolean();
    });
  });

  test("custom actions", () => {
    _testFlow({}, async (ctx) => {
      const res = await ctx.ui.page("page", {
        actions: [
          "thing",
          "another",
          {
            label: "Bingo!",
            value: "bingo",
            mode: "primary",
          },
        ],
        content: [
          ctx.ui.display.header(),
          ctx.ui.inputs.text("name"),
          ctx.ui.inputs.number("age"),
          ctx.ui.inputs.boolean("terms"),
        ],
      });

      expectTypeOf(res.action).toBeString();
      expectTypeOf(res.action).toEqualTypeOf<"thing" | "another" | "bingo">();

      expectTypeOf(res.data.age).toBeNumber();
      expectTypeOf(res.data.name).toBeString();
      expectTypeOf(res.data.terms).toBeBoolean();
    });
  });
});

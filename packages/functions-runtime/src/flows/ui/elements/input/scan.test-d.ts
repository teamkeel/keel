import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("bulk input element", () => {
  test("no config scan", () => {
    testFlow({}, async (ctx) => {
      const bulk = ctx.ui.inputs.scan("scanner");

      const res = await ctx.ui.page("page", {
        content: [bulk],
      });

      expectTypeOf(res.scanner).toBeString();
    });
  });
  test("single scan", () => {
    testFlow({}, async (ctx) => {
      const bulk = ctx.ui.inputs.scan("scanner", {
        mode: "single",
      });

      const res = await ctx.ui.page("page", {
        content: [bulk],
      });

      expectTypeOf(res.scanner).toBeString();
    });
  });
  test("simple mulit scan", () => {
    testFlow({}, async (ctx) => {
      const bulk = ctx.ui.inputs.scan("scanner", {
        mode: "multi",
      });

      const res = await ctx.ui.page("page", {
        content: [bulk],
      });

      expectTypeOf(res.scanner).toBeArray();
      expectTypeOf(res.scanner).toEqualTypeOf<string[]>();
    });
  });
  test("quantity mulit scan", () => {
    testFlow({}, async (ctx) => {
      const bulk = ctx.ui.inputs.scan("scanner", {
        mode: "multi",
        duplicateHandling: "trackQuantity",
      });

      const res = await ctx.ui.page("page", {
        content: [bulk],
      });

      expectTypeOf(res.scanner).toBeArray();
      expectTypeOf(res.scanner).toEqualTypeOf<
        {
          value: string;
          quantity: number;
        }[]
      >;
    });
  });

  test("reject duplicates mulit scan", () => {
    testFlow({}, async (ctx) => {
      const bulk = ctx.ui.inputs.scan("scanner", {
        mode: "multi",
        duplicateHandling: "rejectDuplicates",
      });

      const res = await ctx.ui.page("page", {
        content: [bulk],
      });

      expectTypeOf(res.scanner).toBeArray();
      expectTypeOf(res.scanner).toEqualTypeOf<string[]>();
    });
  });
});

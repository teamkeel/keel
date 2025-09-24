import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("bulk input element", () => {
  test("simple bulk input", () => {
    testFlow({}, async (ctx) => {
      const bulk = ctx.ui.interactive.bulkScan("bulkScan");

      const res = await ctx.ui.page("page", {
        content: [bulk],
      });

      expectTypeOf(res.bulkScan.scans).toBeArray();
      expectTypeOf(res.bulkScan.scans).toEqualTypeOf<string[]>();
    });
  });
  test("quantity bulk input", () => {
    testFlow({}, async (ctx) => {
      type a = Parameters<typeof ctx.ui.interactive.bulkScan>["1"];
      type b = NonNullable<NonNullable<a>["duplicateHandling"]>;

      const bulk = ctx.ui.interactive.bulkScan("bulkScan", {
        duplicateHandling: "trackQuantity",
      });

      const res = await ctx.ui.page("page", {
        content: [bulk],
      });

      expectTypeOf(res.bulkScan.scans).toBeArray();
      expectTypeOf(res.bulkScan.scans).toEqualTypeOf<
        {
          value: string;
          quantity: number;
        }[]
      >;
    });
  });

  test("reject duplicates bulk input", () => {
    testFlow({}, async (ctx) => {
      const bulk = ctx.ui.interactive.bulkScan("bulkScan", {
        duplicateHandling: "rejectDuplicates",
      });

      const res = await ctx.ui.page("page", {
        content: [bulk],
      });

      expectTypeOf(res.bulkScan.scans).toBeArray();
      expectTypeOf(res.bulkScan.scans).toEqualTypeOf<string[]>();
    });
  });
});

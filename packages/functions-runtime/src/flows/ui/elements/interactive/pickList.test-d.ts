import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("pick list element", () => {
  test("pick list types", () => {
    testFlow({}, async (ctx) => {
      const pick = ctx.ui.interactive.pickList("pickList", {
        data: [
          {
            id: "1",
            name: "thing",
            targetQuantity: 10,
            gtins: ["1234567890123"],
          },
          {
            id: "2",
            name: "thing2",
            targetQuantity: 2,
            gtins: ["2397y49y3"],
          },
        ],
        render: (data) => ({
          id: data.id,
          targetQuantity: 1,
          title: data.name,
          barcodes: data.gtins,
        }),
      });

      const res = await ctx.ui.page("page", {
        content: [pick],
      });

      expectTypeOf(res.pickList.items).toBeArray();
      expectTypeOf(res.pickList.items).branded.toEqualTypeOf<
        {
          id: string;
          qty: number;
        }[]
      >;
    });
  });
});

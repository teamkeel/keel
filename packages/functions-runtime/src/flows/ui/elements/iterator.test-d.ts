import { expectTypeOf, test } from "vitest";
import { testFlow, testFlowContext } from "../../testingUtils";

test("iterator element", () => {
  test("all columns", () => {
    testFlow<{}, {}, undefined>({}, async (ctx) => {
      const select = ctx.ui.select.one("selectOne", {
        options: ["1", "2"],
      });

      const it = ctx.ui.iterator("things", {
        content: [
          ctx.ui.display.banner({
            title: "Thing",
            description: "another",
          }),
          ctx.ui.inputs.text("text"),
          select,
        ],
      });

      expectTypeOf(it.name).toBeString();
      expectTypeOf(it.name).toEqualTypeOf<"things">();

      // Check the internal type of content data is correct
      expectTypeOf(it.contentData).toBeArray();
      expectTypeOf(it.contentData[0].text).toBeString();

      const res = await ctx.ui.page("page", {
        content: [it, ctx.ui.inputs.text("text2"), select],
      });

      expectTypeOf(res).toBeObject();
      expectTypeOf(res.text2).toBeString();

      expectTypeOf(res.things).toBeArray();
      expectTypeOf(res.things).toEqualTypeOf<typeof it.contentData>();

      // Inferred types like selection options are the same via a page or in the iterator
      expectTypeOf(res.things[0].selectOne).toEqualTypeOf<
        typeof res.selectOne
      >();
    });
  });
});

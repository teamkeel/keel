import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../testingUtils";

describe("complete", () => {
  test("autoclose and content types", () => {
    testFlow({}, async (ctx) => {
      // Can have content if autoClose is not defined
      ctx.complete({
        title: "title",
        description: "description",
        content: [
          ctx.ui.display.banner({
            title: "title",
            description: "description",
          }),
        ],
        data: {},
      });

      // Can have content and autoClose is false
      ctx.complete({
        title: "title",
        description: "description",
        content: [
          ctx.ui.display.banner({
            title: "title",
            description: "description",
          }),
        ],
        autoClose: false,
        data: {},
      });

      // @ts-expect-error autoClose is not allowed if content is provided
      ctx.complete({
        title: "title",
        description: "description",
        content: [
          ctx.ui.display.banner({
            title: "title",
            description: "description",
          }),
        ],
        autoClose: true,
        data: {},
      });
    });
  });
});

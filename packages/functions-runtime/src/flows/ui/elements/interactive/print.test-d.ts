import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("print element", () => {
  test("pick list types", () => {
    testFlow<
      {},
      {},
      {
        printers: [{ name: "test" }, { name: "test2" }];
      }
    >({}, async (ctx) => {
      const print = ctx.ui.interactive.print({
        jobs: [
          {
            type: "zpl",
            data: "test",
            printer: "test2",
          },
        ],
      });
    });
  });
});

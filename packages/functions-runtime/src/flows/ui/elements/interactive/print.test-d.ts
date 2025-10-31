import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("print element", () => {
  test("printer routing", () => {
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
  test("either data or url must be provided", () => {
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
          },
        ],
      });

      ctx.ui.interactive.print({
        jobs: [
          {
            type: "zpl",
            url: "http://example.com",
          },
        ],
      });

      ctx.ui.interactive.print({
        jobs: [
          {
            type: "rawPdf",
            url: "http://example.com",
          },
        ],
      });

      ctx.ui.interactive.print({
        jobs: [
          // @ts-expect-error - can't have both data and url
          {
            type: "zpl",
            url: "http://example.com",
            data: "test",
          },
        ],
      });

      ctx.ui.interactive.print({
        jobs: [
          // @ts-expect-error - can't have rawPdf with data
          {
            type: "rawPdf",
            data: "test",
          },
        ],
      });

      ctx.ui.interactive.print({
        jobs: [
          // @ts-expect-error - data is required
          {
            type: "zpl",
          },
        ],
      });
    });
  });
});

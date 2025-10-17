import { describe, expectTypeOf, test } from "vitest";
import { testFlow } from "../../../testingUtils";

describe("datepicker input element", () => {
  test("datetime", () => {
    testFlow({}, async (ctx) => {
      const res = await ctx.ui.page("page", {
        content: [
          ctx.ui.inputs.datePicker("date", {
            mode: "dateTime",
            label: "Label",
            pastOnly: false,
          }),
        ],
      });

      expectTypeOf(res.date).toBeString();
    });
  });
});

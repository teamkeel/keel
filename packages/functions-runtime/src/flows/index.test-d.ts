import { test, expect } from "vitest";
import { testFlow } from "./testingUtils";

test("stages types work correctly", () => {
  testFlow(
    {
      title: "my flow",
      stages: [
        "start",
        "middle",
        "end",
        {
          key: "complete",
          name: "All done!",
          initiallyHidden: true,
        },
      ],
    },
    async (ctx) => {
      ctx.step(
        "step",
        {
          stage: "complete",
        },
        async () => {}
      );

      ctx.step(
        "step",
        {
          // @ts-expect-error stage must be one of the keys in the config
          stage: "not a stage",
        },
        async () => {}
      );

      ctx.ui.page("name", {
        stage: "start",
        content: [],
      });

      ctx.ui.page("name", {
        // @ts-expect-error stage must be one of the keys in the config
        stage: "fds",
        content: [],
      });
    }
  );
});

test("json serializable constraint", () => {
  testFlow({}, async (ctx) => {
    await ctx.step("step", async () => {
      return {
        a: 1,
        b: "2",
        c: true,
        d: [1, 2, 3],
        e: {
          f: "g",
          h: 1,
          i: true,
        },
        j: null,
        k: undefined,
        l: BigInt(10),
        m: [{ n: [1, 2, 3] }],
      };
    });
  });
});

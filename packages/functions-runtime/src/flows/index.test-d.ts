import { test } from "vitest";
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

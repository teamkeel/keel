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

  test("autoRestart types", () => {
    type TestInput = { testInput?: string };

    testFlow<{}, TestInput, undefined>({}, async (ctx) => {
      ctx.complete({
        allowRestart: false,
      });

      ctx.complete({
        // @ts-expect-error autoRestart can't just be true if there are inputs
        allowRestart: true,
      });

      ctx.complete({
        allowRestart: {
          inputs: {
            testInput: "test",
          },
          mode: "auto",
          buttonLabel: "test",
        },
      });

      // Test that inputs are optional if there are only optional inputs
      expectTypeOf<
        Parameters<typeof ctx.complete>[0]["allowRestart"]
      >().branded.toEqualTypeOf<
        | {
            inputs?: TestInput;
            mode?: "manual" | "auto";
            buttonLabel?: string;
          }
        | false
        | undefined
      >();
    });

    // Test that inputs are required if there are inputs
    testFlow<{}, { id: string }, undefined>({}, async (ctx) => {
      ctx.complete({
        allowRestart: {
          inputs: {
            id: "test",
          },
        },
      });

      expectTypeOf<
        Parameters<typeof ctx.complete>[0]["allowRestart"]
      >().branded.toEqualTypeOf<
        | {
            inputs: { id: string };
            mode?: "manual" | "auto";
            buttonLabel?: string;
          }
        | false
        | undefined
      >();
    });

    // Flows without inputs
    testFlow<{}, never, undefined>({}, async (ctx) => {
      ctx.complete({
        allowRestart: false,
      });

      ctx.complete({
        allowRestart: true,
      });

      ctx.complete({
        allowRestart: {
          mode: "auto",
        },
      });
    });
  });
});

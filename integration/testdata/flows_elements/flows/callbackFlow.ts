import { CallbackFlow, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default CallbackFlow(config, async (ctx) => {
  const { numberInput, boolInput } = await ctx.ui.page("my page", {
    content: [
      ctx.ui.inputs.number("numberInput", {
        label: "How many numbers?",
        defaultValue: 1,
        onLeave: (callbackInput: number) => {
          return callbackInput * 2;
        },
      }),
      ctx.ui.inputs.boolean("boolInput", {
        label: "True?",
        onLeave: (callbackInput: boolean) => {
          return !callbackInput;
        },
      }),
    ],
  });

  return { numberInput, boolInput };
});

import { CallbackFlow, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default CallbackFlow(config, async (ctx) => {
  const { numberInput } = await ctx.ui.page("my page", {
    content: [
         ctx.ui.inputs.number("numberInput", {
                    label: "How many numbers?",
                    defaultValue: 1,
                    onLeave: (callbackInput) => {
                        return {
                            result: callbackInput.number * 2,
                        }
                    }
                }),
    ]
  });

  return numberInput;
});

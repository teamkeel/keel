import { NoFields, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default NoFields(config, async (ctx, inputs) => {
  await ctx.step("return task entity id", async () => {
    return inputs.entityId;
  });
});

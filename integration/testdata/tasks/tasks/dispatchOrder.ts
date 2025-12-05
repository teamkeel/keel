import { DispatchOrder, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default DispatchOrder(config, async (ctx, inputs) => {});

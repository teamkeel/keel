import { createFlowContext, FlowFunction } from ".";
import { FlowConfig } from ".";

export const testFlowContext = <T extends FlowConfig>(config?: T) =>
  createFlowContext<T, {}, {}, {}>("test-run-id", {}, null, "test-span-id", {
    env: {},
    now: new Date(),
    secrets: {},
    identity: {},
  });

export const testFlow = <const C extends FlowConfig>(
  config: C,
  fn: FlowFunction<C, {}, {}, {}>
) => {
  return { config, fn };
};

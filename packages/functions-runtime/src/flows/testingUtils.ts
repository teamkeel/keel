import { createFlowContext, FlowFunction } from ".";
import { FlowConfig } from ".";

export const _testFlowContext = <T extends FlowConfig>(config?: T) =>
  createFlowContext<T>("test-run-id", {}, "test-span-id");

export const _testFlow = <const C extends FlowConfig>(
  config: C,
  fn: FlowFunction<C, {}>
) => {
  return { config, fn };
};

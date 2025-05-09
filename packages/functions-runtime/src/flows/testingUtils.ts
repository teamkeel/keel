import { createFlowContext, FlowFunction } from ".";
import { FlowConfig } from ".";

export const testFlowContext = <T extends FlowConfig>(config?: T) =>
  createFlowContext<T>("test-run-id", {}, "test-span-id");

export const testFlow = <const C extends FlowConfig>(
  config: C,
  fn: FlowFunction<C, {}>
) => {
  return { config, fn };
};

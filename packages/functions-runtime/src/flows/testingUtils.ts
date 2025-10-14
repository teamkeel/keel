import { createFlowContext, FlowFunction, NullableHardware } from ".";
import { FlowConfig } from ".";

export const testFlowContext = <T extends FlowConfig>(config?: T) =>
  createFlowContext<T, {}, {}, {}, { testInput: string }, undefined>(
    "test-run-id",
    {},
    null,
    null,
    null,
    "test-span-id",
    {
      env: {},
      now: new Date(),
      secrets: {},
      identity: {},
    }
  );

export const testFlow = <
  const C extends FlowConfig,
  I,
  H extends NullableHardware,
>(
  config: C,
  fn: FlowFunction<C, {}, {}, {}, I, H>
) => {
  return { config, fn };
};

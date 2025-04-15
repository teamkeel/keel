import { UI } from "./ui";

type FlowInputs = Record<string, any>;

interface StepContext<C extends FlowConfig, I extends {}> {
  inputs: I;
  step: <R = any>(
    name: string,
    fn: (step: { skip(): void }) => Promise<R>
  ) => Promise<R> & {
    catch: (errorHandler: (err: Error) => Promise<void> | void) => Promise<any>;
  };
  ui: UI<C>;
}

export type FlowFunction<C extends FlowConfig = {}, I extends {} = never> = (
  context: StepContext<C, I>
) => any;

export interface FlowConfig {
  stages?: StageConfig[];
  title?: string;
  description?: string;
}

type StageConfig =
  | string
  | {
      key: string;
      name: string;
      description?: string;
      initiallyHidden?: boolean;
    };

// Function overloads
export function flow<I extends FlowInputs, const C extends FlowConfig>(
  flowName: string,
  config: C,
  flow: FlowFunction<C, I>
): (inputs: I) => any;
export function flow<
  I extends FlowInputs = {},
  const C extends FlowConfig = {},
>(flowName: string, flow: FlowFunction<C, I>): (inputs: I) => any;

// ****************************
// Mock implementation (to be replaced)
// ****************************

export function flow<
  I extends FlowInputs = {},
  const C extends FlowConfig = {},
>(
  flowName: string,
  configOrFlow: C | FlowFunction<C, I>,
  flowFunction?: FlowFunction<C, I>
) {
  const config = typeof configOrFlow === "function" ? undefined : configOrFlow;
  const flow =
    typeof configOrFlow === "function" ? configOrFlow : flowFunction!;

  return async (inputs: I) => {
    const ctx = createStepContext<C, I>(inputs);
    return flow(ctx);
  };
}

function createStepContext<C extends FlowConfig, I extends FlowInputs = {}>(
  inputs: I
): StepContext<C, I> {
  return {
    inputs: {} as I,
    step: <T = any>(
      name: string,
      fn: (ctx: { skip: () => void }) => Promise<T>
    ) => {
      const stepPromise = fn({} as any);
      const stepWithCatch = Object.assign(stepPromise, {
        catch: async (errorHandler: (err: Error) => Promise<void> | void) => {
          try {
            return await stepPromise;
          } catch (err) {
            await errorHandler(err as Error);
            throw err;
          }
        },
      });
      return stepWithCatch;
    },
    ui: {} as any,
  };
}

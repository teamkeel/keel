import { UI } from "./ui";

interface FlowInputs {
  email: string;
  name: string;
  userId: string;
  [key: string]: any; // Allow for additional inputs
}

interface StepContext<C extends FlowConfig> {
  step: <R = any>(
    name: string,
    fn: (step: { skip(): void }) => Promise<R>
  ) => Promise<R> & {
    catch: (errorHandler: (err: Error) => Promise<void> | void) => Promise<any>;
  };
  ui: UI<C>;
}

export type FlowFunction<C extends FlowConfig = {}> = (
  context: FlowContext<C>
) => any;

// Update FlowContext to include stage type parameter
interface FlowContext<C extends FlowConfig> {
  ctx: StepContext<C>;
  inputs: FlowInputs;
}

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
export function flow<const C extends FlowConfig>(
  flowName: string,
  config: C,
  flow: FlowFunction<C>
): (inputs: FlowInputs) => any;
export function flow(
  flowName: string,
  flow: FlowFunction
): (inputs: FlowInputs) => any;

// ****************************
// Mock implementation (to be replaced)
// ****************************

export function flow<C extends FlowConfig>(
  flowName: string,
  configOrFlow: C,
  flowFunction?: FlowFunction<C>
) {
  const config = typeof configOrFlow === "function" ? undefined : configOrFlow;
  const flow =
    typeof configOrFlow === "function" ? configOrFlow : flowFunction!;

  return async (inputs: FlowInputs) => {
    const ctx = createStepContext<C>();
    return flow({ ctx, inputs });
  };
}

function createStepContext<C extends FlowConfig>(): StepContext<C> {
  return {
    step: <T = any>(
      name: string,
      fn: (ctx: StepContext<C> & { skip: () => void }) => Promise<T>
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

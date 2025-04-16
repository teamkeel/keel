import { UI } from "./ui";
import { useDatabase } from "../database";
import { FlowDisrupt } from "../StepRunner";
export { UI }

const STEP_STATUS = {
  NEW: "NEW",
  COMPLETED: "COMPLETED",
  FAILED: "FAILED",
};

const STEP_TYPE = {
  FUNCTION: "FUNCTION",
  IO: "IO",
  DELAY: "DELAY",
};

const defaultOpts = {
  maxRetries: 5,
  timeoutInMs: 60000,
};

type FlowInputs = Record<string, any>;

interface StepContext<C extends FlowConfig> {
  //inputs: I;
  step: <R = any>(
    name: string,
    fn: () => Promise<R>
  ) => Promise<R> & {
    catch: (errorHandler: (err: Error) => Promise<void> | void) => Promise<any>;
  };
  ui: UI<C>;
}

export type FlowFunction<C extends FlowConfig = {}> = (
  context: StepContext<C>
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
// export function flow<I extends FlowInputs, const C extends FlowConfig>(
//   flowName: string,
//   config: C,
//   flow: FlowFunction<C>
// ): (inputs: I) => any;

// export function flow<
//   I extends FlowInputs = {},
//   const C extends FlowConfig = {},
// >(flowName: string, flow: FlowFunction<C>): (inputs: I) => any;

// ****************************
// Mock implementation (to be replaced)
// ****************************

// export function flow<
//   const C extends FlowConfig = {},
// >(
//   flowName: string,
//   configOrFlow: C | FlowFunction<C>,
//   flowFunction?: FlowFunction<C>
// ) {
//   const config = typeof configOrFlow === "function" ? undefined : configOrFlow;
//   const flow =
//     typeof configOrFlow === "function" ? configOrFlow : flowFunction!;

//   return async (inputs: any) => {
//     const ctx = createStepContext<C>();
//     return flow(ctx);
//   };
// }

type Opts = {
  maxRetries?: number;
  timeoutInMs?: number;
}

export function createStepContext<C extends FlowConfig>(runId: string): StepContext<C> {
  return {
   // inputs: {} as I,
    step: async <T = any>(
      name: string,
      fn: () => Promise<T>,
      opts?: Opts
    ) => {

      const db = useDatabase();

      // First check if we already have a result for this step
      const completed = await db
        .selectFrom("keel_flow_step")
        .where("run_id", "=", runId)
        .where("name", "=", name)
        .where("status", "=", STEP_STATUS.COMPLETED)
        .selectAll()
        .executeTakeFirst();

      if (completed) {
        return completed.value;
      }

      // The step hasn't yet run successfully, so we need to create a NEW run
      const step = await db
        .insertInto("keel_flow_step")
        .values({
          run_id: runId,
          name: name,
          status: STEP_STATUS.NEW,
          type: STEP_TYPE.FUNCTION,
          maxRetries: opts?.maxRetries ?? defaultOpts.maxRetries,
          timeoutInMs: opts?.timeoutInMs ?? defaultOpts.timeoutInMs,
        })
        .returningAll()
        .executeTakeFirst();

      let outcome = STEP_STATUS.COMPLETED;

      let result = null;
      try {
        result = await  withTimeout(fn(), step.timeoutInMs);
      } catch (e) {
        outcome = STEP_STATUS.FAILED;
      }

      // Very crudely store the result in the database
      await db
        .updateTable("keel_flow_step")
        .set({
          status: outcome,
          value: JSON.stringify(result),
        })
        .where("id", "=", step.id)
        .returningAll()
        .executeTakeFirst();

      throw new FlowDisrupt();

      // const stepPromise = fn({} as any);
      // const stepWithCatch = Object.assign(stepPromise, {
      //   catch: async (errorHandler: (err: Error) => Promise<void> | void) => {
      //     try {
      //       return await stepPromise;
      //     } catch (err) {
      //       await errorHandler(err as Error);
      //       throw err;
      //     }
      //   },
      // });
      // return stepWithCatch;
    },
    ui: {
       page: async (
        options: any
      ) => {

        const db = useDatabase();

         // First check if we already have a result for this step
        const step = await db
          .selectFrom("keel_flow_step")
          .where("run_id", "=", runId)
          .where("name", "=", options.title)
          .selectAll()
          .executeTakeFirst();


        if (!step) {
          // The step hasn't yet run successfully, so we need to create a NEW run
          const step = await db
            .insertInto("keel_flow_step")
            .values({
              run_id: runId,
              name: options.title,
              status: "RUNNING",
              type: "UI",
              maxRetries: 3,
              timeoutInMs: 1000,
            })
            .returningAll()
            .executeTakeFirst();

          throw options;
        }

        console.log(step);


        switch (step.status) {
          case "RUNNING":
            throw options;
          case "COMPLETED":
            return step.data;
        }


        //return options;
          
        // Get this step from the database and determine next move:
        //  - If the step is RUNNING and there is no data, then return flow UI structure.  We are still waiting on UI data.
        //  - If the step is RUNNING and there is data, then run the validation functions. If these all pass, then update the step to COMPLETED.
        //  - If the step is COMPLETED, then return the data.

      },
    } as any,
  };
}

function wait(milliseconds: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

function withTimeout<T>(promiseFn: Promise<T>, timeout: number): Promise<T> {
  return Promise.race([
    promiseFn,
    wait(timeout).then(() => {
      throw new Error(`flow times out after ${timeout}ms`);
    }),
  ]);
}


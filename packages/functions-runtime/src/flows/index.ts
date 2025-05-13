import { UI } from "./ui";
import { useDatabase } from "../database";
import { textInput } from "./ui/elements/input/text";
import { numberInput } from "./ui/elements/input/number";
import { divider } from "./ui/elements/display/divider";
import { booleanInput } from "./ui/elements/input/boolean";
import { markdown } from "./ui/elements/display/markdown";
import { table } from "./ui/elements/display/table";
import { selectOne } from "./ui/elements/select/one";
import { page, UiPage } from "./ui/page";
import {
  StepCreatedDisrupt,
  UIRenderDisrupt,
  ExhuastedRetriesDisrupt,
} from "./disrupts";
import { banner } from "./ui/elements/display/banner";
import { image } from "./ui/elements/display/image";
import { code } from "./ui/elements/display/code";
import { grid } from "./ui/elements/display/grid";
import { list } from "./ui/elements/display/list";
import { header } from "./ui/elements/display/header";

const enum STEP_STATUS {
  NEW = "NEW",
  RUNNING = "RUNNING",
  PENDING = "PENDING",
  COMPLETED = "COMPLETED",
  FAILED = "FAILED",
}

const enum STEP_TYPE {
  FUNCTION = "FUNCTION",
  UI = "UI",
  DELAY = "DELAY",
}

const defaultOpts = {
  maxRetries: 5,
  timeoutInMs: 60000,
};

export interface FlowContext<C extends FlowConfig> {
  step: Step<C>;
  ui: UI<C>;
}

// Steps can only return values that can be serialized to JSON and then
// deserialized back to the same object/value that represents the type.
// i.e. the string, number and boolean primitives, and arrays of them and objects made up of them.
type JsonSerializable =
  | string
  | number
  | boolean
  | null
  | JsonSerializable[]
  | { [key: string]: JsonSerializable };

export type Step<C extends FlowConfig> = {
  <R extends JsonSerializable | void>(
    name: string,
    options: {
      stage?: ExtractStageKeys<C>;
      maxRetries?: number;
      timeoutInMs?: number;
    },
    fn: () => Promise<R> & {
      catch: (
        errorHandler: (err: Error) => Promise<void> | void
      ) => Promise<any>;
    }
  ): Promise<R>;
  <R extends JsonSerializable | void>(
    name: string,
    fn: () => Promise<R> & {
      catch: (
        errorHandler: (err: Error) => Promise<void> | void
      ) => Promise<any>;
    }
  ): Promise<R>;
};

type StepFunction<R> = () => Promise<R> & {
  catch: (errorHandler: (err: Error) => Promise<void> | void) => Promise<any>;
};

export interface FlowConfig {
  stages?: StageConfig[];
  title?: string;
  description?: string;
}

export type FlowFunction<C extends FlowConfig, I extends any = {}> = (
  ctx: FlowContext<C>,
  inputs: I
) => Promise<void>;

// Extract the stage keys from the flow config supporting either a string or an object with a key property
export type ExtractStageKeys<T extends FlowConfig> = T extends {
  stages: infer S;
}
  ? S extends ReadonlyArray<infer U>
    ? U extends string
      ? U
      : U extends { key: infer K extends string }
      ? K
      : never
    : never
  : never;

type StageConfig =
  | string
  | {
      key: string;
      name: string;
      description?: string;
      initiallyHidden?: boolean;
    };

export function createFlowContext<C extends FlowConfig>(
  runId: string,
  data: any,
  spanId: string
): FlowContext<C> {
  return {
    step: async (name, optionsOrFn, fn?) => {
      // We need to check the type of the arguments due to the step function being overloaded
      const options = typeof optionsOrFn === "function" ? {} : optionsOrFn;
      const actualFn = (
        typeof optionsOrFn === "function" ? optionsOrFn : fn!
      ) as StepFunction<any>;

      const db = useDatabase();

      // First check if we already have a result for this step
      const past = await db
        .selectFrom("keel_flow_step")
        .where("run_id", "=", runId)
        .where("name", "=", name)
        .selectAll()
        .execute();

      const newSteps = past.filter((step) => step.status === STEP_STATUS.NEW);
      const completedSteps = past.filter(
        (step) => step.status === STEP_STATUS.COMPLETED
      );
      const failedSteps = past.filter(
        (step) => step.status === STEP_STATUS.FAILED
      );

      if (newSteps.length > 1) {
        throw new Error("Multiple NEW steps found for the same step");
      }

      if (completedSteps.length > 1) {
        throw new Error("Multiple completed steps found for the same step");
      }

      if (completedSteps.length > 1 && newSteps.length > 1) {
        throw new Error(
          "Multiple completed and new steps found for the same step"
        );
      }

      if (completedSteps.length === 1) {
        return completedSteps[0].value;
      }

      // Do we have a NEW step waiting to be run?
      if (newSteps.length === 1) {
        let result = null;
        await db
          .updateTable("keel_flow_step")
          .set({
            startTime: new Date(),
          })
          .where("id", "=", newSteps[0].id)
          .returningAll()
          .executeTakeFirst();

        try {
          result = await withTimeout(
            actualFn(),
            options.timeoutInMs ?? defaultOpts.timeoutInMs
          );
        } catch (e) {
          await db
            .updateTable("keel_flow_step")
            .set({
              status: STEP_STATUS.FAILED,
              spanId: spanId,
              endTime: new Date(),
              error: e instanceof Error ? e.message : "An error occurred",
            })
            .where("id", "=", newSteps[0].id)
            .returningAll()
            .executeTakeFirst();

          if (
            failedSteps.length + 1 >=
            (options.maxRetries ?? defaultOpts.maxRetries)
          ) {
            throw new ExhuastedRetriesDisrupt();
          }

          // If we have retries left, create a new step
          await db
            .insertInto("keel_flow_step")
            .values({
              run_id: runId,
              name: name,
              stage: options.stage,
              status: STEP_STATUS.NEW,
              type: STEP_TYPE.FUNCTION,
            })
            .returningAll()
            .executeTakeFirst();

          throw new StepCreatedDisrupt();
        }

        // Store the result in the database
        await db
          .updateTable("keel_flow_step")
          .set({
            status: STEP_STATUS.COMPLETED,
            value: JSON.stringify(result),
            spanId: spanId,
            endTime: new Date(),
          })
          .where("id", "=", newSteps[0].id)
          .returningAll()
          .executeTakeFirst();

        return result;
      }

      // The step hasn't yet run successfully, so we need to create a NEW run
      await db
        .insertInto("keel_flow_step")
        .values({
          run_id: runId,
          name: name,
          stage: options.stage,
          status: STEP_STATUS.NEW,
          type: STEP_TYPE.FUNCTION,
        })
        .returningAll()
        .executeTakeFirst();

      throw new StepCreatedDisrupt();

      // TODO: Incorporate when we have support error handling
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
      page: (async (name, options) => {
        const db = useDatabase();

        // First check if this step exists
        let step = await db
          .selectFrom("keel_flow_step")
          .where("run_id", "=", runId)
          .where("name", "=", name)
          .selectAll()
          .executeTakeFirst();

        // If this step has already been completed, return the values. Steps are only ever run to completion once.
        if (step && step.status === STEP_STATUS.COMPLETED) {
          return step.value;
        }

        if (!step) {
          // The step hasn't yet run so we create a new the step with state PENDING.
          step = await db
            .insertInto("keel_flow_step")
            .values({
              run_id: runId,
              name: name,
              stage: options.stage,
              status: STEP_STATUS.PENDING,
              type: STEP_TYPE.UI,
              startTime: new Date(),
            })
            .returningAll()
            .executeTakeFirst();

          // If no data has been passed in, render the UI by disrupting the step with UIRenderDisrupt.
          throw new UIRenderDisrupt(step?.id, page(options));
        }

        if (!data) {
          // If no data has been passed in, render the UI by disrupting the step with UIRenderDisrupt.
          throw new UIRenderDisrupt(step?.id, page(options));
        }

        // TODO: Validate the data! If not valid, throw a UIRenderDisrupt along with the validation errors.

        // If the data has been passed in and is valid, persist the data and mark the step as COMPLETED, and then return the data.
        await db
          .updateTable("keel_flow_step")
          .set({
            status: STEP_STATUS.COMPLETED,
            value: JSON.stringify(data),
            spanId: spanId,
            endTime: new Date(),
          })
          .where("id", "=", step.id)
          .returningAll()
          .executeTakeFirst();

        return data;
      }) as UiPage<C>,
      inputs: {
        text: textInput as any,
        number: numberInput as any,
        boolean: booleanInput as any,
      },
      display: {
        divider: divider as any,
        markdown: markdown as any,
        table: table as any,
        header: header as any,
        banner: banner as any,
        image: image as any,
        code: code as any,
        grid: grid as any,
        list: list as any,
      },
      select: {
        one: selectOne as any,
      },
    },
  };
}

function wait(milliseconds: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

function withTimeout<T>(promiseFn: Promise<T>, timeout: number): Promise<T> {
  return Promise.race([
    promiseFn,
    wait(timeout).then(() => {
      throw new Error(`Step function timed out after ${timeout}ms`);
    }),
  ]);
}

export { UI };

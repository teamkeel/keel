import { UI } from "./ui";
import { Complete, CompleteOptions } from "./ui/complete";
import { useDatabase } from "../database";
import { textInput } from "./ui/elements/input/text";
import { numberInput } from "./ui/elements/input/number";
import { divider } from "./ui/elements/display/divider";
import { booleanInput } from "./ui/elements/input/boolean";
import { markdown } from "./ui/elements/display/markdown";
import { table } from "./ui/elements/display/table";
import { selectOne } from "./ui/elements/select/one";
import { page, callbackFn, UiPage } from "./ui/page";
import {
  StepCreatedDisrupt,
  UIRenderDisrupt,
  ExhuastedRetriesDisrupt,
  CallbackDisrupt,
} from "./disrupts";
import { banner } from "./ui/elements/display/banner";
import { image } from "./ui/elements/display/image";
import { code } from "./ui/elements/display/code";
import { grid } from "./ui/elements/display/grid";
import { list } from "./ui/elements/display/list";
import { header } from "./ui/elements/display/header";
import { keyValue } from "./ui/elements/display/keyValue";
import { selectTable } from "./ui/elements/select/table";
import { dataGridInput } from "./ui/elements/input/dataGrid";
import { iterator } from "./ui/elements/iterator";
import { print } from "./ui/elements/interactive/print";
import { pickList } from "./ui/elements/interactive/pickList";
import { NonRetriableError } from "./errors";
import { scan } from "./ui/elements/input/scan";

export const enum STEP_STATUS {
  NEW = "NEW",
  RUNNING = "RUNNING",
  PENDING = "PENDING",
  COMPLETED = "COMPLETED",
  FAILED = "FAILED",
}

export const enum STEP_TYPE {
  FUNCTION = "FUNCTION",
  UI = "UI",
  DELAY = "DELAY",
  COMPLETE = "COMPLETE",
}

/** A function used to calculate the delay between attempting a retry. The returned value is the number of ms of delay. */
type RetryPolicyFn = (retry: number) => number;

/**
 * Returns a linear backoff retry delay.
 * @param intervalS duration in seconds before the first retry. The second retry will double it, the third triple it and so on.
 */
export const RetryBackoffLinear = (intervalS: number): RetryPolicyFn => {
  return (retry: number) => retry * intervalS * 1000;
};

/**
 * Retuns a constant retry delay.
 * @param intervalS duration in seconds between retries.
 */
export const RetryConstant = (intervalS: number): RetryPolicyFn => {
  return (retry: number) => (retry > 0 ? intervalS * 1000 : 0);
};

/**
 * Returns an exponential backoff retry delay.
 * @param intervalS the base duration in seconds.
 */
export const RetryBackoffExponential = (intervalS: number): RetryPolicyFn => {
  return (retry: number) => {
    if (retry < 1) {
      return 0;
    }
    return Math.pow(intervalS, retry) * 1000;
  };
};

const defaultOpts = {
  retries: 4,
  timeout: 60000,
};

export interface FlowContext<C extends FlowConfig, E, S, Id, I> {
  // Defines a function step that will be run in the flow.
  step: Step<C>;
  // Defines a UI step that will be run in the flow.
  ui: UI<C>;
  complete: Complete<C, I>;
  env: E;
  now: Date;
  secrets: S;
  identity: Id;
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

type StepOptions<C extends FlowConfig> = {
  stage?: ExtractStageKeys<C>;
  /** Number of times to retry the step after it fails. Defaults to 4. */
  retries?: number;
  /** Function to calculate the delay before retrying this step. By default steps will be retried immediately. */
  retryPolicy?: RetryPolicyFn;
  /** Maximum time in milliseconds to wait for the step to complete. Defaults to 60000 (1 minute). */
  timeout?: number;
  /** A function to call if the step fails after it exhausts all retries. */
  onFailure?: () => Promise<void> | void;
};

export type Step<C extends FlowConfig> = {
  <R extends JsonSerializable | void>(
    /** The unique name of this step. */
    name: string,
    /** Configuration options for the step. */
    options: StepOptions<C>,
    /** The step function to execute. */
    fn: StepFunction<C, R>
  ): Promise<R>;
  <R extends JsonSerializable | void>(
    /** The unique name of this step. */
    name: string,
    /** The step function to execute. */
    fn: StepFunction<C, R>
  ): Promise<R>;
};

type StepArgs<C extends FlowConfig> = {
  attempt: number;
  stepOptions: StepOptions<C>;
};

type StepFunction<C extends FlowConfig, R> = (args: StepArgs<C>) => Promise<R>;

export interface FlowConfig {
  /** The stages to organise the steps in the flow. */
  stages?: StageConfig[];
  /** The title of the flow as shown in the Console. */
  title?: string;
  /** The description of the flow as shown in the Console. */
  description?: string;
}

// What is returned as the config to the API
export interface FlowConfigAPI {
  stages?: StageConfigObject[];
  title: string;
  description?: string;
}

export type FlowFunction<C extends FlowConfig, E, S, Id, I = undefined> = (
  ctx: FlowContext<C, E, S, Id, I>,
  inputs: I
) => Promise<CompleteOptions<C, I> | any | void>;

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

type StageConfigObject = {
  /** The unique key of the stage. */
  key: string;
  /** The name of the stage as shown in the Console. */
  name: string;
  /** The description of the stage as shown in the Console. */
  description?: string;
  /** Whether the stage is initially hidden in the Console. */
  initiallyHidden?: boolean;
};

type StageConfig = string | StageConfigObject;

export function createFlowContext<C extends FlowConfig, E, S, Id, I>(
  runId: string,
  data: any,
  action: string | null,
  callback: string | null,
  element: string | null,
  spanId: string,
  ctx: {
    env: E;
    now: Date;
    secrets: S;
    identity: Id;
  }
): FlowContext<C, E, S, Id, I> {
  // Track step and page names to prevent duplicates
  const usedNames = new Set<string>();

  return {
    identity: ctx.identity,
    env: ctx.env,
    now: ctx.now,
    secrets: ctx.secrets,
    complete: (options) => {
      return {
        __type: "ui.complete",
        ...options,
      };
    },
    step: async (name, optionsOrFn, fn?) => {
      // We need to check the type of the arguments due to the step function being overloaded
      const options = typeof optionsOrFn === "function" ? {} : optionsOrFn;
      const actualFn = (
        typeof optionsOrFn === "function" ? optionsOrFn : fn!
      ) as StepFunction<C, any>;

      options.retries = options.retries ?? defaultOpts.retries;
      options.timeout = options.timeout ?? defaultOpts.timeout;

      const db = useDatabase();

      // Check for duplicate step names
      if (usedNames.has(name)) {
        await db
          .insertInto("keel.flow_step")
          .values({
            run_id: runId,
            name: name,
            stage: options.stage,
            status: STEP_STATUS.FAILED,
            type: STEP_TYPE.FUNCTION,
            error: `Duplicate step name: ${name}`,
            startTime: new Date(),
            endTime: new Date(),
          })
          .returningAll()
          .executeTakeFirst();

        throw new Error(`Duplicate step name: ${name}`);
      }
      usedNames.add(name);

      // First check if we already have a result for this step
      const past = await db
        .selectFrom("keel.flow_step")
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
          .updateTable("keel.flow_step")
          .set({
            startTime: new Date(),
          })
          .where("id", "=", newSteps[0].id)
          .returningAll()
          .executeTakeFirst();

        try {
          const stepArgs: StepArgs<C> = {
            attempt: failedSteps.length + 1,
            stepOptions: options,
          };

          result = await withTimeout(actualFn(stepArgs), options.timeout);
        } catch (e) {
          await db
            .updateTable("keel.flow_step")
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
            failedSteps.length >= options.retries ||
            e instanceof NonRetriableError
          ) {
            if (options.onFailure) {
              await options.onFailure!();
            }
            throw new ExhuastedRetriesDisrupt();
          }

          // If we have retries left, create a new step
          await db
            .insertInto("keel.flow_step")
            .values({
              run_id: runId,
              name: name,
              stage: options.stage,
              status: STEP_STATUS.NEW,
              type: STEP_TYPE.FUNCTION,
            })
            .returningAll()
            .executeTakeFirst();

          throw new StepCreatedDisrupt(
            options.retryPolicy
              ? new Date(
                  Date.now() + options.retryPolicy(failedSteps.length + 1)
                )
              : undefined
          );
        }

        // Store the result in the database
        await db
          .updateTable("keel.flow_step")
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
        .insertInto("keel.flow_step")
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
    },
    ui: {
      page: (async (name, options) => {
        const db = useDatabase();

        const isCallback = element && callback;

        // Check for duplicate step names
        if (usedNames.has(name)) {
          await db
            .insertInto("keel.flow_step")
            .values({
              run_id: runId,
              name: name,
              stage: options.stage,
              status: STEP_STATUS.FAILED,
              type: STEP_TYPE.UI,
              error: `Duplicate step name: ${name}`,
              startTime: new Date(),
              endTime: new Date(),
            })
            .returningAll()
            .executeTakeFirst();

          throw new Error(`Duplicate step name: ${name}`);
        }
        usedNames.add(name);

        // First check if this step exists
        let step = await db
          .selectFrom("keel.flow_step")
          .where("run_id", "=", runId)
          .where("name", "=", name)
          .selectAll()
          .executeTakeFirst();

        // If this step has already been completed, return the values. Steps are only ever run to completion once.
        if (step && step.status === STEP_STATUS.COMPLETED) {
          if (step.action) {
            return { data: step.value, action: step.action };
          }
          return step.value;
        }

        if (!step) {
          // The step hasn't yet run so we create a new the step with state PENDING.
          step = await db
            .insertInto("keel.flow_step")
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

          // We now render the UI by disrupting the step with UIRenderDisrupt.
          throw new UIRenderDisrupt(
            step?.id,
            (await page(options, null, null)).page
          );
        }

        if (isCallback) {
          // we now need to resolve a UI callback.
          try {
            const response = await callbackFn(
              options.content,
              element,
              callback,
              data
            );
            throw new CallbackDisrupt(response, false);
          } catch (e) {
            if (e instanceof CallbackDisrupt) {
              throw e;
            }

            throw new CallbackDisrupt(
              e instanceof Error ? e.message : `An error occurred`,
              true
            );
          }
        }

        if (!data) {
          // If no data has been passed in, render the UI by disrupting the step with UIRenderDisrupt.
          throw new UIRenderDisrupt(
            step?.id,
            (await page(options, null, null)).page
          );
        }

        try {
          const p = await page(options, data, action);

          if (p.hasValidationErrors) {
            throw new UIRenderDisrupt(step?.id, p.page);
          }
        } catch (e) {
          if (e instanceof UIRenderDisrupt) {
            throw e;
          }

          await db
            .updateTable("keel.flow_step")
            .set({
              status: STEP_STATUS.FAILED,
              spanId: spanId,
              endTime: new Date(),
              error: e instanceof Error ? e.message : "An error occurred",
            })
            .where("id", "=", step?.id)
            .returningAll()
            .executeTakeFirst();

          throw e;
        }

        // If the data has been passed in and is valid, persist the data (and action if applicable) and mark the step as COMPLETED, and then return the data.
        await db
          .updateTable("keel.flow_step")
          .set({
            status: STEP_STATUS.COMPLETED,
            value: JSON.stringify(data),
            action: action,
            spanId: spanId,
            endTime: new Date(),
          })
          .where("id", "=", step.id)
          .returningAll()
          .executeTakeFirst();

        if (action) {
          return { data, action };
        }
        return data;
      }) as UiPage<C>,
      inputs: {
        text: textInput as any,
        number: numberInput as any,
        boolean: booleanInput as any,
        dataGrid: dataGridInput as any,
        scan: scan as any,
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
        keyValue: keyValue as any,
      },
      select: {
        one: selectOne as any,
        table: selectTable as any,
      },
      iterator: iterator as any,
      interactive: {
        print: print as any,
        pickList: pickList as any,
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

export { UI, NonRetriableError };

import { UI } from "./ui";
import { Complete, CompleteOptions } from "./ui/complete";
import { useDatabase } from "../database";
import {
  withSpan,
  KEEL_INTERNAL_ATTR,
  KEEL_INTERNAL_CHILDREN,
} from "../tracing";
import * as opentelemetry from "@opentelemetry/api";
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
import { datePickerInput } from "./ui/elements/input/datePicker";
import { fileInput } from "./ui/elements/input/file";
import { iterator } from "./ui/elements/iterator";
import { print } from "./ui/elements/interactive/print";
import { pickList } from "./ui/elements/interactive/pickList";
import { NonRetriableError } from "./errors";
import { scan } from "./ui/elements/input/scan";
import { file } from "./ui/elements/display/file";
import { transformRichDataTypes } from "../parsing";

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

export interface FlowContext<
  C extends FlowConfig,
  E,
  S,
  Id,
  I,
  H extends NullableHardware,
> {
  // Defines a function step that will be run in the flow.
  step: Step<C>;
  // Defines a UI step that will be run in the flow.
  ui: UI<C, H>;
  complete: Complete<C, I>;
  env: E;
  now: Date;
  secrets: S;
  identity: Id;
}

export type NullableHardware = Hardware | undefined;

export type Hardware = {
  printers: Printer[];
};
export interface Printer {
  name: string;
}

// Steps can only return values that can be serialized to JSON and then
// deserialized back to the same object/value that represents the type.
// i.e. the string, number and boolean primitives, and arrays of them and objects made up of them.
type JsonSerializable =
  | string
  | number
  | boolean
  | null
  | Date
  | JsonSerializable[]
  | { [key: string]: JsonSerializable }
  | Map<string, JsonSerializable>;

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

export type FlowFunction<
  C extends FlowConfig,
  E,
  S,
  Id,
  I = undefined,
  H extends NullableHardware = undefined,
> = (
  ctx: FlowContext<C, E, S, Id, I, H>,
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

export function createFlowContext<
  C extends FlowConfig,
  E,
  S,
  Id,
  I,
  H extends NullableHardware,
>(
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
): FlowContext<C, E, S, Id, I, H> {
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
      return withSpan(`Step - ${name}`, async (span: opentelemetry.Span) => {
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
          // step already executed, so this tracing span is internal
          span.setAttribute(KEEL_INTERNAL_ATTR, KEEL_INTERNAL_CHILDREN);
          return deserializeValue(completedSteps[0].value);
        }

        // Do we have a NEW step waiting to be run?
        if (newSteps.length === 1) {
          let result: any = null;
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
              value: serializeValue(result),
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

        // step was just created, so this tracing span should be internal
        span.setAttribute(KEEL_INTERNAL_ATTR, KEEL_INTERNAL_CHILDREN);
        throw new StepCreatedDisrupt();
      });
    },
    ui: {
      page: (async (name, options) => {
        return withSpan(`Page - ${name}`, async (span: opentelemetry.Span) => {
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
            // page already completed, so we're marking this span as internal alongside it's children
            span.setAttribute(KEEL_INTERNAL_ATTR, KEEL_INTERNAL_CHILDREN);

            const parsedData = transformRichDataTypes(step.value);

            if (step.action) {
              // When actions are present, the flow always returns { data, action }
              // so we need to maintain that structure when returning from DB
              return { data: parsedData, action: step.action };
            }
            // Without actions, just return the data directly
            return parsedData;
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

            span.setAttribute("rendered", true);

            // We now render the UI by disrupting the step with UIRenderDisrupt.
            throw new UIRenderDisrupt(
              step?.id,
              (await page(options, null, null)).page
            );
          }

          if (isCallback) {
            span.setAttribute("callback", callback);

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
              value: serializeValue(data),
              action: action,
              spanId: spanId,
              endTime: new Date(),
            })
            .where("id", "=", step.id)
            .returningAll()
            .executeTakeFirst();

          const parsedData = transformRichDataTypes(data);

          // Only return the { data, action } wrapper when actions are defined
          if (action) {
            return { data: parsedData, action };
          }
          return parsedData;
        });
      }) as UiPage<C>,
      inputs: {
        text: textInput as any,
        number: numberInput as any,
        boolean: booleanInput as any,
        dataGrid: dataGridInput as any,
        datePicker: datePickerInput as any,
        scan: scan as any,
        file: fileInput as any,
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
        file: file as any,
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

// Custom JSON replacer to handle Map and Date serialization
function jsonReplacer(key: string, value: any): any {
  if (value instanceof Map) {
    return Object.fromEntries(value);
  }
  // Date objects are automatically serialized to ISO strings by JSON.stringify
  // so we don't need special handling here
  return value;
}

// Helper to convert Maps to plain objects before stringification
function serializeValue(value: any): string {
  // Handle the case where the root value is a Map
  if (value instanceof Map) {
    return JSON.stringify(Object.fromEntries(value), jsonReplacer);
  }
  // Date objects are handled natively by JSON.stringify (converted to ISO strings)
  // Model objects from database queries are plain objects and serialize normally
  return JSON.stringify(value, jsonReplacer);
}

// Helper to deserialize values and convert ISO date strings back to Date objects
function deserializeValue(value: any): any {
  // ISO 8601 date string pattern
  const isoDatePattern = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z$/;

  if (value === null || value === undefined) {
    return value;
  }

  // If it's a string that looks like an ISO date, convert to Date
  if (typeof value === 'string' && isoDatePattern.test(value)) {
    return new Date(value);
  }

  // If it's an array, recursively deserialize each element
  if (Array.isArray(value)) {
    return value.map(deserializeValue);
  }

  // If it's an object, recursively deserialize each property
  if (typeof value === 'object') {
    const result: any = {};
    for (const key in value) {
      if (value.hasOwnProperty(key)) {
        result[key] = deserializeValue(value[key]);
      }
    }
    return result;
  }

  // For primitives, return as-is
  return value;
}

export { UI, NonRetriableError };

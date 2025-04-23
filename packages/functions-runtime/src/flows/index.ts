import { UI } from "./ui";
import { useDatabase } from "../database";
import { textInput } from "./ui/elements/input/text";
import { numberInput } from "./ui/elements/input/number";
import { divider } from "./ui/elements/display/divider";
import { booleanInput } from "./ui/elements/input/boolean";
import { markdown } from "./ui/elements/display/markdown";
import { table } from "./ui/elements/display/table";
import { selectOne } from "./ui/elements/select/single";
import { UiPage } from "./ui/page";
import {
  StepCompletedDisrupt,
  StepErrorDisrupt,
  UIRenderDisrupt,
} from "./disrupts";

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

export interface StepContext<C extends FlowConfig> {
  step: <R = any>(
    name: string,
    fn: () => Promise<R>,
    opts?: Opts
  ) => Promise<R> & {
    catch: (errorHandler: (err: Error) => Promise<void> | void) => Promise<any>;
  };
  ui: UI<C>;
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

type Opts = {
  maxRetries?: number;
  timeoutInMs?: number;
};

export function createStepContext<C extends FlowConfig>(
  runId: string
): StepContext<C> {
  return {
    step: async <T = any>(name: string, fn: () => Promise<T>, opts?: Opts) => {
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

      let result = null;
      try {
        result = await withTimeout(fn(), step.timeoutInMs);
      } catch (e) {
        await db
          .updateTable("keel_flow_step")
          .set({
            status: STEP_STATUS.FAILED,
            // TODO: store error message
          })
          .where("id", "=", step.id)
          .returningAll()
          .executeTakeFirst();

        throw new StepErrorDisrupt(e instanceof Error ? e.message : "an error occurred");
      }

      // Very crudely store the result in the database
      await db
        .updateTable("keel_flow_step")
        .set({
          status: STEP_STATUS.COMPLETED,
          value: JSON.stringify(result),
        })
        .where("id", "=", step.id)
        .returningAll()
        .executeTakeFirst();

      throw new StepCompletedDisrupt();

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
      page: (async (page) => {
        const db = useDatabase();

        // First check if we already have a result for this step
        let step = await db
          .selectFrom("keel_flow_step")
          .where("run_id", "=", runId)
          .where("name", "=", page.title)
          .selectAll()
          .executeTakeFirst();

        if (!step) {
          // The step hasn't yet run successfully, so we need to create a NEW run
          step = await db
            .insertInto("keel_flow_step")
            .values({
              run_id: runId,
              name: page.title,
              status: STEP_STATUS.PENDING,
              type: STEP_TYPE.UI,
              maxRetries: 3,
              timeoutInMs: 1000,
            })
            .returningAll()
            .executeTakeFirst();
        }

        switch (step.status) {
          case STEP_STATUS.PENDING:
            throw new UIRenderDisrupt(step.id, page);
          case STEP_STATUS.COMPLETED:
            return step.value;
          default:
            throw new StepErrorDisrupt("ui step failed");
        }
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
      },
      select: {
        single: selectOne as any,
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
      throw new Error(`flow times out after ${timeout}ms`);
    }),
  ]);
}

export { UI };

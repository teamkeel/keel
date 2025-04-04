const { useDatabase } = require("./database");

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

// This is a special type that is thrown to disrupt the execution of a flow
class FlowDisrupt {
  constructor() {}
}

class StepRunner {
  constructor(runId) {
    this.runId = runId;
  }

  async run(name, fn, opts) {
    const db = useDatabase();
    console.log(opts);
    // First check if we already have a result for this step
    const completed = await db
      .selectFrom("keel_flow_step")
      .where("run_id", "=", this.runId)
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
        run_id: this.runId,
        name: name,
        status: STEP_STATUS.NEW,
        type: STEP_TYPE.FUNCTION,
        maxRetries: opts?.maxRetries ?? 3,
        timeoutInMs: opts?.timeoutInMs ?? 500,
      })
      .returningAll()
      .executeTakeFirst();

    let outcome = STEP_STATUS.COMPLETED;

    let result = null;
    try {
      result = await withTimeout(fn(), { timeout: step.timeoutInMs });
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
  }
}

function wait(milliseconds) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

export function withTimeout(promiseFn, opts) {
  const timeOut = opts?.timeout ?? 2500;

  return Promise.race([
    promiseFn,
    wait(timeOut).then(() => {
      throw new Error(`flow times out after ${timeout}ms`);
    }),
  ]);
}

module.exports = { StepRunner, FlowDisrupt };

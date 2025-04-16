const {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} = require("json-rpc-2.0");
const { createDatabaseClient } = require("./database");
const { errorToJSONRPCResponse, RuntimeErrors } = require("./errors");
const opentelemetry = require("@opentelemetry/api");
const { withSpan } = require("./tracing");
const { tryExecuteFlow } = require("./tryExecuteFlow");
const { parseInputs } = require("./parsing");
const { FlowDisrupt } = require("./StepRunner");
const { createStepContext } = require("./flows");

async function handleFlow(request, config) {
  // Try to extract trace context from caller
  const activeContext = opentelemetry.propagation.extract(
    opentelemetry.context.active(),
    request.meta?.tracing
  );

  // Run the whole request with the extracted context
  return opentelemetry.context.with(activeContext, () => {
    // Wrapping span for the whole request
    return withSpan(request.method, async (span) => {
      let db = null;

      const runId = request.meta?.runId;

      try {
        if (!runId) {
          throw new Error("no runId provided");
        }

        const { createFlowContextAPI, flows } = config;

        if (!(request.method in flows)) {
          const message = `no corresponding flow found for '${request.method}'`;
          span.setStatus({
            code: opentelemetry.SpanStatusCode.ERROR,
            message: message,
          });
          return createJSONRPCErrorResponse(
            request.id,
            JSONRPCErrorCode.MethodNotFound,
            message
          );
        }

        const ctx = createStepContext(request.meta.runId);

        // // The ctx argument passed into the flow function.
        // const ctx = createFlowContextAPI({
        //   meta: request.meta,
        // });

        db = createDatabaseClient({
          connString: request.meta?.secrets?.KEEL_DB_CONN,
        });

        const flowRun = await db
          .selectFrom("keel_flow_run")
          .where("id", "=", runId)
          .selectAll()
          .executeTakeFirst();

        if (!flowRun) {
          throw new Error("no flow run found");
        }

        const flowFunction = flows[request.method];

        await tryExecuteFlow(db, async () => {
          // parse request params to convert objects into rich field types (e.g. InlineFile)
          const inputs = parseInputs(flowRun.input);

          return flowFunction(ctx, inputs);
        });

        // If we reach this point, then we know the entire flow completed successfully
        // TODO: Send FlowRunUpdated event with run_completed = true
        return createJSONRPCSuccessResponse(request.id, {
          runId: runId,
          runCompleted: true,
        });
      } catch (e) {
        // If the flow is disrupted, then we know that a step either completed successfully or failed
        if (e instanceof FlowDisrupt) {
          // TODO: Send FlowRunUpdated event with run_completed = false
          return createJSONRPCSuccessResponse(request.id, {
            runId: runId,
            runCompleted: false,
          });
        }

        if (e instanceof Error) {
          span.recordException(e);
          span.setStatus({
            code: opentelemetry.SpanStatusCode.ERROR,
            message: e.message,
          });
          return errorToJSONRPCResponse(request, e);
        }

        const message = JSON.stringify(e);

        span.setStatus({
          code: opentelemetry.SpanStatusCode.ERROR,
          message: message,
        });

        return createJSONRPCErrorResponse(
          request.id,
          RuntimeErrors.UnknownError,
          message
        );
      } finally {
        if (db) {
          await db.destroy();
        }
      }
    });
  });
}

module.exports = {
  handleFlow,
  RuntimeErrors,
};

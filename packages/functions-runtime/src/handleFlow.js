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
const { createFlowContext } = require("./flows");
const {
  StepCompletedDisrupt,
  StepErrorDisrupt,
  UIRenderDisrupt,
} = require("./flows/disrupts");

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
      let flowConfig = null;
      const runId = request.meta?.runId;

      try {
        if (!runId) {
          throw new Error("no runId provided");
        }

        const { flows } = config;

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

        const ctx = createFlowContext(
          request.meta.runId,
          request.meta.data,
          span.spanContext().spanId
        );

        const flowFunction = flows[request.method].fn;
        flowConfig = flows[request.method].config;

        await tryExecuteFlow(db, async () => {
          // parse request params to convert objects into rich field types (e.g. InlineFile)
          const inputs = parseInputs(flowRun.input);

          return flowFunction(ctx, inputs);
        });

        // If we reach this point, then we know the entire flow completed successfully
        return createJSONRPCSuccessResponse(request.id, {
          runId: runId,
          runCompleted: true,
          config: flowConfig,
        });
      } catch (e) {
        // The flow is disrupted by a function step completion
        if (e instanceof StepCompletedDisrupt) {
          return createJSONRPCSuccessResponse(request.id, {
            runId: runId,
            runCompleted: false,
            config: flowConfig,
          });
        }

        // The flow is disrupted by a pending UI step
        if (e instanceof UIRenderDisrupt) {
          return createJSONRPCSuccessResponse(request.id, {
            runId: runId,
            stepId: e.stepId,
            config: flowConfig,
            ui: e.contents,
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

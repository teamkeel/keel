import {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} from "json-rpc-2.0";
import { createDatabaseClient } from "./database";
import { errorToJSONRPCResponse, RuntimeErrors } from "./errors";
import * as opentelemetry from "@opentelemetry/api";
import { withSpan } from "./tracing";
import { tryExecuteFlow } from "./tryExecuteFlow";
import { parseInputs } from "./parsing";
import { createFlowContext } from "./flows";
import {
  StepCreatedDisrupt,
  UIRenderDisrupt,
  ExhuastedRetriesDisrupt,
} from "./flows/disrupts";

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
          .selectFrom("keel.flow_run")
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
        // The flow is disrupted as a new step has been created
        if (e instanceof StepCreatedDisrupt) {
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

        span.recordException(e);
        span.setStatus({
          code: opentelemetry.SpanStatusCode.ERROR,
          message: e.message,
        });

        // The flow has failed due to exhausted step retries
        if (e instanceof ExhuastedRetriesDisrupt) {
          return createJSONRPCSuccessResponse(request.id, {
            runId: runId,
            runCompleted: true,
            error: "flow failed due to exhausted step retries",
            config: flowConfig,
          });
        }

        return createJSONRPCErrorResponse(
          request.id,
          JSONRPCErrorCode.InternalError,
          e.message
        );
      } finally {
        if (db) {
          await db.destroy();
        }
      }
    });
  });
}

export { handleFlow, RuntimeErrors };

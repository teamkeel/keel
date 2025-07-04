import {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} from "json-rpc-2.0";
import { createDatabaseClient } from "./database";
import { errorToJSONRPCResponse, RuntimeErrors } from "./errors";
import * as opentelemetry from "@opentelemetry/api";
import { withSpan } from "./tracing";
import { tryExecuteJob } from "./tryExecuteJob";
import { parseInputs } from "./parsing";
import { PROTO_ACTION_TYPES } from "./consts";

// Generic handler function that is agnostic to runtime environment (local or lambda)
// to execute a job function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
async function handleJob(request, config) {
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

      try {
        const { createJobContextAPI, jobs } = config;

        if (!jobs[request.method]) {
          const message = `job '${request.method}' does not exist or has not been implemented`;
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

        // The ctx argument passed into the job function.
        const ctx = createJobContextAPI({
          meta: request.meta,
        });

        const permitted =
          request.meta && request.meta.permissionState.status === "granted"
            ? true
            : null;

        db = createDatabaseClient({
          connString: request.meta?.secrets?.KEEL_DB_CONN,
        });
        const jobFunction = jobs[request.method];
        const actionType = PROTO_ACTION_TYPES.JOB;

        const functionConfig = jobFunction?.config ?? {};

        await tryExecuteJob(
          { request, permitted, db, actionType, functionConfig },
          async () => {
            // parse request params to convert objects into rich field types (e.g. InlineFile)
            const inputs = parseInputs(request.params);

            // Return the job function to the containing tryExecuteJob block
            return jobFunction(ctx, inputs);
          }
        );

        return createJSONRPCSuccessResponse(request.id, null);
      } catch (e) {
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

export { handleJob, RuntimeErrors };

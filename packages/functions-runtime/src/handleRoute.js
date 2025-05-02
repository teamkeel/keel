import {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} from "json-rpc-2.0";
import { createDatabaseClient, withDatabase } from "./database";
import { withAuditContext } from "./auditing";
import { errorToJSONRPCResponse, RuntimeErrors } from "./errors";
import * as opentelemetry from "@opentelemetry/api";
import { withSpan } from "./tracing";

async function handleRoute(request, config) {
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
        const { createContextAPI, functions } = config;

        if (!(request.method in functions)) {
          const message = `no route function found for '${request.method}'`;
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

        // For route functions context doesn't need request headers or the response object as this is handled by
        // params and the function response respectively
        const {
          headers,
          response: __,
          ...ctx
        } = createContextAPI({
          responseHeaders: new Headers(),
          meta: request.meta,
        });

        // Add request headers to params
        request.params.headers = headers;

        db = createDatabaseClient({
          connString: request.meta?.secrets?.KEEL_DB_CONN,
        });
        const routeHandler = functions[request.method];

        const result = await withDatabase(db, false, () => {
          return withAuditContext(request, () => {
            return routeHandler(request.params, ctx);
          });
        });

        if (result instanceof Error) {
          span.recordException(result);
          span.setStatus({
            code: opentelemetry.SpanStatusCode.ERROR,
            message: result.message,
          });
          return errorToJSONRPCResponse(request, result);
        }

        const response = createJSONRPCSuccessResponse(request.id, result);

        return response;
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

export { handleRoute, RuntimeErrors };

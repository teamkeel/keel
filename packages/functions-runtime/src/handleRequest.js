import {
  createJSONRPCErrorResponse,
  createJSONRPCSuccessResponse,
  JSONRPCErrorCode,
} from "json-rpc-2.0";
import { createDatabaseClient } from "./database";
import { tryExecuteFunction } from "./tryExecuteFunction";
import { errorToJSONRPCResponse, RuntimeErrors } from "./errors";
import * as opentelemetry from "@opentelemetry/api";
import { withSpan } from "./tracing";
import { parseInputs, parseOutputs } from "./parsing";

// Generic handler function that is agnostic to runtime environment (local or lambda)
// to execute a custom function based on the contents of a jsonrpc-2.0 payload object.
// To read more about jsonrpc request and response shapes, please read https://www.jsonrpc.org/specification
async function handleRequest(request, config) {
  // Try to extract trace context from caller
  const activeContext = opentelemetry.propagation.extract(
    opentelemetry.context.active(),
    request.meta?.tracing
  );

  if (process.env.KEEL_LOG_LEVEL == "debug") {
    console.log(request);
  }

  // Run the whole request with the extracted context
  return opentelemetry.context.with(activeContext, () => {
    // Wrapping span for the whole request
    return withSpan(request.method, async (span) => {
      let db = null;

      try {
        const { createContextAPI, functions, permissionFns, actionTypes } =
          config;

        if (!functions[request.method]) {
          const message = `function '${request.method}' does not exist or has not been implemented`;
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

        // headers reference passed to custom function where object data can be modified
        const headers = new Headers();

        // The ctx argument passed into the custom function.
        const ctx = createContextAPI({
          responseHeaders: headers,
          meta: request.meta,
        });

        // The Go runtime does *some* permissions checks up front before the request reaches
        // this method, so we pass in a permissionState object on the request.meta object that
        // indicates whether a call to a custom function has already been authorised
        const permitted =
          request.meta && request.meta.permissionState.status === "granted"
            ? true
            : null;

        db = createDatabaseClient({
          connString: request.meta?.secrets?.KEEL_DB_CONN,
        });
        const customFunction = functions[request.method];
        const actionType = actionTypes[request.method];

        const functionConfig = customFunction?.config ?? {};

        const result = await tryExecuteFunction(
          {
            request,
            ctx,
            permitted,
            db,
            permissionFns,
            actionType,
            functionConfig,
          },
          async () => {
            // parse request params to convert objects into rich field types (e.g. InlineFile)
            const inputs = parseInputs(request.params);

            // Return the custom function to the containing tryExecuteFunction block
            // Once the custom function is called, tryExecuteFunction will check the schema's permission rules to see if it can continue committing
            // the transaction to the db. If a permission rule is violated, any changes made inside the transaction are rolled back.
            const result = await customFunction(ctx, inputs);

            return parseOutputs(result);
          }
        );

        if (result instanceof Error) {
          span.recordException(result);
          span.setStatus({
            code: opentelemetry.SpanStatusCode.ERROR,
            message: result.message,
          });
          return errorToJSONRPCResponse(request, result);
        }

        const response = createJSONRPCSuccessResponse(request.id, result);

        const responseHeaders = {};
        for (const pair of headers.entries()) {
          responseHeaders[pair[0]] = pair[1].split(", ");
        }
        response.meta = {
          headers: responseHeaders,
          status: ctx.response.status,
        };

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

export { handleRequest, RuntimeErrors };

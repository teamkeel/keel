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
import { createFlowContext, FlowConfig, STEP_STATUS, STEP_TYPE } from "./flows";
import {
  CompleteOptions,
  UiCompleteApiResponse,
  complete,
} from "./flows/ui/complete";

import {
  StepCreatedDisrupt,
  UIRenderDisrupt,
  ExhuastedRetriesDisrupt,
} from "./flows/disrupts";
import { sentenceCase } from "change-case";

async function handleFlow(request: any, config: any) {
  // Try to extract trace context from caller
  const activeContext = opentelemetry.propagation.extract(
    opentelemetry.context.active(),
    request.meta?.tracing
  );

  // Run the whole request with the extracted context
  return opentelemetry.context.with(activeContext, () => {
    // Wrapping span for the whole request
    return withSpan(request.method, async (span: any) => {
      let db = null;
      let flowConfig = null;
      const runId = request.meta?.runId;

      try {
        if (!runId) {
          throw new Error("no runId provided");
        }

        const { flows, createFlowContextAPI } = config;

        if (!flows[request.method]) {
          const message = `flow '${request.method}' does not exist or has not been implemented`;
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

        const ctx = createFlowContext(
          request.meta.runId,
          request.meta.data,
          request.meta.action,
          span.spanContext().spanId,
          createFlowContextAPI({
            meta: request.meta,
          })
        );

        const flowFunction = flows[request.method].fn;

        // Normalise the flow config
        const rawFlowConfig: FlowConfig = flows[request.method].config;
        flowConfig = {
          ...rawFlowConfig,
          title: rawFlowConfig.title || sentenceCase(request.method || "flow"),
          stages: rawFlowConfig.stages?.map((stage) => {
            if (typeof stage === "string") {
              return {
                key: stage,
                name: stage,
              };
            }
            return stage;
          }),
        };

        // parse request params to convert objects into rich field types (e.g. InlineFile)
        const inputs = parseInputs(request.meta?.inputs);

        let response: CompleteOptions<FlowConfig> | any | void = undefined;

        try {
          response = await tryExecuteFlow(db, async () => {
            return flowFunction(ctx, inputs);
          });
        } catch (e) {
          // The flow is disrupted as a new step has been created
          if (e instanceof StepCreatedDisrupt) {
            return createJSONRPCSuccessResponse(request.id, {
              runId: runId,
              runCompleted: false,
              config: flowConfig,
              executeAfter: e.executeAfter,
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
            message: e instanceof Error ? e.message : "unknown error",
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

          return createJSONRPCSuccessResponse(request.id, {
            runId: runId,
            runCompleted: true,
            error: e instanceof Error ? e.message : "unknown error",
            config: flowConfig,
          });
        }

        let ui: UiCompleteApiResponse | null = null;
        let data: any = null;

        // TODO: this is not a thorough enough check for the response type
        if (
          response &&
          typeof response == "object" &&
          "__type" in response &&
          response.__type === "ui.complete"
        ) {
          ui = await complete(response);

          const completeStep = await db
            .selectFrom("keel.flow_step")
            .where("run_id", "=", runId)
            .where("type", "=", STEP_TYPE.COMPLETE)
            .selectAll()
            .executeTakeFirst();

          if (!completeStep) {
            await db
              .insertInto("keel.flow_step")
              .values({
                run_id: runId,
                name: "",
                stage: response.stage,
                status: STEP_STATUS.COMPLETED,
                type: STEP_TYPE.COMPLETE,
                startTime: new Date(),
                endTime: new Date(),
                ui: JSON.stringify(ui),
              })
              .returningAll()
              .executeTakeFirst();
          }

          data = response.data;
        } else if (response) {
          data = response;
        }

        // If we reach this point, then we know the entire flow completed successfully
        return createJSONRPCSuccessResponse(request.id, {
          runId: runId,
          runCompleted: true,
          data: data,
          config: flowConfig,
        });
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

export { handleFlow, RuntimeErrors };

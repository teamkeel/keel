import { AsyncLocalStorage } from "async_hooks";
import TraceParent from "traceparent";
import { sql, SelectionNode } from "kysely";

const auditContextStorage = new AsyncLocalStorage();

// withAuditContext creates the audit context from the runtime request body
// and sets it to in AsyncLocalStorage so that this data is available to the
// ModelAPI during the execution of actions, jobs and subscribers.
async function withAuditContext(request, cb) {
  let audit = {};

  if (request.meta?.identity) {
    audit.identityId = request.meta.identity.id;
  }
  if (request.meta?.tracing?.traceparent) {
    audit.traceId = TraceParent.fromString(
      request.meta.tracing.traceparent
    )?.traceId;
  }

  return await auditContextStorage.run(audit, () => {
    return cb();
  });
}

// getAuditContext retrieves the audit context from AsyncLocalStorage.
function getAuditContext() {
  let auditStore = auditContextStorage.getStore();
  return {
    identityId: auditStore?.identityId,
    traceId: auditStore?.traceId,
  };
}

// AuditContextPlugin is a Kysely plugin which ensures that the audit context data
// is written to Postgres configuration parameters in the same execution as a query.
// It does this by calling the set_identity_id() and set_trace_id() functions as a
// clause in the returning statement. It then subsequently drops these from the actual result.
// This ensures that these parameters are set when the tables' AFTER trigger function executes.
class AuditContextPlugin {
  constructor() {
    this.identityIdAlias = "__keel_identity_id";
    this.traceIdAlias = "__keel_trace_id";
  }

  // Appends set_identity_id() and set_trace_id() function calls to the returning statement
  // of INSERT, UPDATE and DELETE operations.
  transformQuery(args) {
    switch (args.node.kind) {
      case "InsertQueryNode":
      case "UpdateQueryNode":
      case "DeleteQueryNode":
        // Represents a RETURNING clause in a SQL statement.
        const returning = {
          kind: "ReturningNode",
          selections: [],
        };

        // If the query already has a selection, then append it.
        if (args.node.returning) {
          returning.selections.push(...args.node.returning.selections);
        }

        // Retrieve the audit context from async storage.
        const audit = getAuditContext();

        if (audit.identityId) {
          const rawNode = sql`set_identity_id(${audit.identityId})`
            .as(this.identityIdAlias)
            .toOperationNode();

          returning.selections.push(SelectionNode.create(rawNode));
        }

        if (audit.traceId) {
          const rawNode = sql`set_trace_id(${audit.traceId})`
            .as(this.traceIdAlias)
            .toOperationNode();

          returning.selections.push(SelectionNode.create(rawNode));
        }

        return {
          ...args.node,
          returning: returning,
        };
    }

    return {
      ...args.node,
    };
  }

  // Drops the set_identity_id() and set_trace_id() fields from the result.
  transformResult(args) {
    if (args.result?.rows) {
      for (let i = 0; i < args.result.rows.length; i++) {
        delete args.result.rows[i][this.identityIdAlias];
        delete args.result.rows[i][this.traceIdAlias];
      }
    }

    return args.result;
  }
}

export { withAuditContext, getAuditContext, AuditContextPlugin };

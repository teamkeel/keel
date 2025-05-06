import { createJSONRPCErrorResponse } from "json-rpc-2.0";

const RuntimeErrors = {
  // Catchall error type for unhandled execution errors during custom function
  UnknownError: -32001,
  // DatabaseError represents any error at pg level that isn't handled explicitly below
  DatabaseError: -32002,
  // No result returned from custom function by user
  NoResultError: -32003,
  // When trying to delete/update a non existent record in the db
  RecordNotFoundError: -32004,
  ForeignKeyConstraintError: -32005,
  NotNullConstraintError: -32006,
  UniqueConstraintError: -32007,
  PermissionError: -32008,
  BadRequestError: -32009,
};

// Error presets
class PermissionError extends Error {}

class DatabaseError extends Error {
  constructor(error) {
    super(error.message);
    this.error = error;
  }
}

class NotFoundError extends Error {
  errorCode = RuntimeErrors.RecordNotFoundError;
  constructor(message) {
    super(message); // Default message is handled by the runtime for consistency with built in actions
  }
}

class BadRequestError extends Error {
  errorCode = RuntimeErrors.BadRequestError;
  constructor(message = "bad request") {
    super(message);
  }
}

class UnknownError extends Error {
  errorCode = RuntimeErrors.UnknownError;
  constructor(message = "unknown error") {
    super(message);
  }
}

const ErrorPresets = {
  NotFound: NotFoundError,
  BadRequest: BadRequestError,
  Unknown: UnknownError,
};

// errorToJSONRPCResponse transforms a JavaScript Error instance (or derivative) into a valid JSONRPC response object to pass back to the Keel runtime.
function errorToJSONRPCResponse(request, e) {
  switch (e.constructor.name) {
    case "PermissionError":
      return createJSONRPCErrorResponse(
        request.id,
        RuntimeErrors.PermissionError,
        e.message
      );
    // Any error thrown in the ModelAPI class is
    // wrapped in a DatabaseError in order to differentiate 'our code' vs the user's own code.
    case "NoResultError":
      return createJSONRPCErrorResponse(
        request.id,

        // to be matched to https://github.com/teamkeel/keel/blob/e3115ffe381bfc371d4f45bbf96a15072a994ce5/runtime/actions/update.go#L54-L54
        RuntimeErrors.RecordNotFoundError,
        "" // Don't pass on the message as we want to normalise these at the runtime layer but still support custom messages in other NotFound errors
      );
    case "DatabaseError":
      let err = e;

      // If wrapped error then unwrap
      if (e instanceof DatabaseError) {
        err = e.error;
      }

      if (err.constructor.name == "NoResultError") {
        return createJSONRPCErrorResponse(
          request.id,

          // to be matched to https://github.com/teamkeel/keel/blob/e3115ffe381bfc371d4f45bbf96a15072a994ce5/runtime/actions/update.go#L54-L54
          RuntimeErrors.RecordNotFoundError,
          "" // Don't pass on the message as we want to normalise these at the runtime layer but still support custom messages in other NotFound errors
        );
      }

      // if the error contains 'code' then assume it has other pg error message keys
      // todo: make this more ironclad.
      // when using lib-pq, should match https://github.com/brianc/node-postgres/blob/master/packages/pg-protocol/src/parser.ts#L371-L386
      if ("code" in err) {
        const { code, detail, table } = err;

        let rpcErrorCode, column, value;
        const [col, val] = parseKeyMessage(err.detail);
        column = col;
        value = val;

        switch (code) {
          case "23502":
            rpcErrorCode = RuntimeErrors.NotNullConstraintError;
            column = err.column;
            break;
          case "23503":
            rpcErrorCode = RuntimeErrors.ForeignKeyConstraintError;
            break;
          case "23505":
            rpcErrorCode = RuntimeErrors.UniqueConstraintError;
            break;
          default:
            rpcErrorCode = RuntimeErrors.DatabaseError;
            break;
        }

        return createJSONRPCErrorResponse(request.id, rpcErrorCode, e.message, {
          table,
          column,
          code,
          detail,
          value,
        });
      }

      // we don't know what it is, but it's something else
      return createJSONRPCErrorResponse(
        request.id,
        RuntimeErrors.DatabaseError,
        e.message
      );
    default:
      // Use the errorCode in the error if we have some from a preset
      return createJSONRPCErrorResponse(
        request.id,
        e.errorCode ?? RuntimeErrors.UnknownError,
        e.message
      );
  }
}

// example data:
// Key (author_id)=(fake) is not present in table "author".
const keyMessagePattern = /\Key\s[(](.*)[)][=][(](.*)[)]/;
const parseKeyMessage = (msg) => {
  const [, col, value] = keyMessagePattern.exec(msg) || [];

  return [col, value];
};

export {
  errorToJSONRPCResponse,
  RuntimeErrors,
  DatabaseError,
  PermissionError,
  ErrorPresets,
};

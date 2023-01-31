const { DatabaseError } = require("./ModelAPI");
const { createJSONRPCErrorResponse, JSONRPCErrorCode } = require("json-rpc-2.0");

const RuntimeErrors = {
  UnknownError: -32001,
  DatabaseError: -32002,
};

function errorToJSONRPCResponse(request, e) {
  // we want to switch on instanceof but there is no way to do that in js, so best to check the constructor class of the error

  // todo: fuzzy matching on postgres errors from both rds-data-api and pg-protocol
  switch(e.constructor) {
    case SyntaxError:
      return createJSONRPCErrorResponse(
        request.id,
        JSONRPCErrorCode.InternalError,
        e.message,
        {
          stack: e.stack
        },
      );
    // Unhandled promise rejections in any of the instance methods of the ModelAPI are caught by a wrapping fn and errors are wrapped in a DatabaseError.
    case DatabaseError:
      return createJSONRPCErrorResponse(
        request.id,
        RuntimeErrors.DatabaseError,
        "No result",
      );
    default:
      return createJSONRPCErrorResponse(
        request.id,
        RuntimeErrors.UnknownError,
        e.message,
        {
          stack: e.stack,
        },
      );
  }
}

module.exports = {
  errorToJSONRPCResponse,
  RuntimeErrors,
}
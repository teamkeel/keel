const { DatabaseError } = require("./ModelAPI");
const { createJSONRPCErrorResponse } = require("json-rpc-2.0");

const RuntimeErrors = {
  UnknownError: -32001,
  DatabaseError: -32002,
};

// transforms a JavaScript Error instance (or derivative) into a valid JSONRPC response object
// to pass back to the Keel runtime
function errorToJSONRPCResponse(request, e) {
  // we want to switch on instanceof but there is no way to do that in js, so best to check the constructor class of the error

  // todo: fuzzy matching on postgres errors from both rds-data-api and pg-protocol
  switch (e.constructor) {
    case DatabaseError:
      return createJSONRPCErrorResponse(
        request.id,
        RuntimeErrors.DatabaseError,
        e.message
      );
    default:
      return createJSONRPCErrorResponse(
        request.id,
        RuntimeErrors.UnknownError,
        e.message
      );
  }
}

module.exports = {
  errorToJSONRPCResponse,
  RuntimeErrors,
};

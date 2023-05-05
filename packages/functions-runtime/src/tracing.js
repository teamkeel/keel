const opentelemetry = require("@opentelemetry/api");

const serviceName = "customerCustomFunctions";

function withSpan(name, fn) {
  const tracer = opentelemetry.trace.getTracer(serviceName);

  return tracer.startActiveSpan(name, async (span) => {
    try {
      // await the thing (this means we can use try/catch)
      return await fn(span);
    } catch (err) {
      // record any errors
      span.recordException(err);
      span.setStatus({
        code: opentelemetry.SpanStatusCode.ERROR,
        message: err.message,
      });
      // re-throw the error
      throw err;
    } finally {
      // make sure the span is ended
      span.end();
    }
  });
}

module.exports = {
  serviceName,
  withSpan,
};

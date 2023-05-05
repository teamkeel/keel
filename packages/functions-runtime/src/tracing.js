const opentelemetry = require("@opentelemetry/api");

function withSpan(name, fn) {
  return getTracer().startActiveSpan(name, async (span) => {
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

function init() {
  if (!globalThis.fetch.patched) {
    const originalFetch = globalThis.fetch;

    globalThis.fetch = async (...args) => {
      return withSpan("fetch", async (span) => {
        const url = new URL(
          args[0] instanceof Request ? args[0].url : String(args[0])
        );
        span.setAttribute("http.url", url.toString());
        const scheme = url.protocol.replace(":", "");
        span.setAttribute("http.scheme", scheme);

        const options = args[0] instanceof Request ? args[0] : args[1] || {};
        const method = (options.method || "GET").toUpperCase();
        span.setAttribute("http.method", method);

        const res = await originalFetch(...args);
        span.setAttribute("http.status", res.status);
        span.setAttribute("http.status_text", res.statusText);
      });
    };
    globalThis.fetch.patched = true;
  }
}

function getTracer() {
  return opentelemetry.trace.getTracer("functions");
}

module.exports = {
  getTracer,
  withSpan,
  init,
};

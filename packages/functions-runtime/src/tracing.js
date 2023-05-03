const opentelemetry = require("@opentelemetry/api");

const serviceName = "customerCustomFunctions";

function asyncFunction(name, asyncFunc, attributes) {
  return function () {
    const tracer = opentelemetry.trace.getTracer(serviceName);
    return tracer.startActiveSpan(name, (span) => {
      if (attributes) {
        for (let key of Object.keys(attributes)) {
          span.setAttribute(key, attributes[key]);
        }
      }
      return asyncFunc(span, ...arguments).finally(() => {
        span.end();
      });
    });
  };
}

function promise(name, asyncFunc, attributes) {
  return asyncFunction(name, asyncFunc, attributes)();
}

module.exports = {
  asyncFunction,
  promise,
};

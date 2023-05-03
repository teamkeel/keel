import { expect, test, beforeEach } from "vitest";
import tracing from "./tracing";
import { NodeTracerProvider, Span } from "@opentelemetry/sdk-trace-node";

let spanEvents = [];
const provider = new NodeTracerProvider({});
provider.addSpanProcessor({
  forceFlush() {
    return Promise.resolve();
  },
  onStart(span, parentContext) {
    spanEvents.push({ event: "onStart", span, parentContext });
  },
  onEnd(span) {
    spanEvents.push({ event: "onEnd", span });
  },
  shutdown() {
    return Promise.resolve();
  },
});
provider.register();

beforeEach(() => {
  spanEvents = [];
});

test("asyncFunction arguments passed", async () => {
  let f = tracing.asyncFunction("name", async function () {
    return [...arguments];
  });
  expect(await f()).length(1);
  const result = await f("a", "b", "c");
  expect(result).length(4);
  // drop first element, which is the span
  result.shift();
  expect(result).toEqual(["a", "b", "c"]);
});

test("asyncFunction span time", async () => {
  let started = false;
  let ended = false;
  const waitTimeMillis = 100;
  let f = tracing.asyncFunction("name", async function () {
    started = true;
    await new Promise((resolve) => setTimeout(resolve, waitTimeMillis));
    ended = true;
    return {};
  });

  const resultPromise = f();
  expect(started).toEqual(true);
  expect(ended).toEqual(false);
  expect(spanEvents).length(1);

  await resultPromise;
  expect(started).toEqual(true);
  expect(ended).toEqual(true);
  expect(spanEvents.map((e) => e.event)).toEqual(["onStart", "onEnd"]);
  const spanDuration = spanEvents.pop().span._duration.pop();
  const waitTimeNanos = waitTimeMillis * 1000 * 1000;
  expect(spanDuration).toBeGreaterThan(waitTimeNanos);
});

test("asyncFunction span ends even on error", async () => {
  let started = false;
  let f = tracing.asyncFunction("name", async function () {
    started = true;
    throw "err";
  });

  const resultPromise = f();
  expect(started).toEqual(true);
  expect(spanEvents.map((e) => e.event)).toEqual(["onStart"]);

  try {
    await resultPromise;
    // previous line should have an error thrown
    expect(true).toEqual(false);
  } catch (e) {
    expect(e).toEqual("err");
    expect(spanEvents.map((e) => e.event)).toEqual(["onStart", "onEnd"]);
  }
});

test("asyncFunction attributes", async () => {
  const p = () => Promise.resolve();
  await tracing.asyncFunction("name", p)();
  expect(spanEvents.pop().span.attributes).toEqual({});
  await tracing.asyncFunction("name", p, {})();
  expect(spanEvents.pop().span.attributes).toEqual({});
  await tracing.asyncFunction("name", p, {
    a: 1,
    b: 2,
  })();
  expect(spanEvents.pop().span.attributes).toEqual({
    a: 1,
    b: 2,
  });
});

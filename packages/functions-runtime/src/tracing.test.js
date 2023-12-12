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
  tracing.init();
  spanEvents = [];
});

test("withSpan span time", async () => {
  const waitTimeMillis = 100;
  await tracing.withSpan("name", async () => {
    await new Promise((resolve) => setTimeout(resolve, waitTimeMillis));
  });

  expect(spanEvents.map((e) => e.event)).toEqual(["onStart", "onEnd"]);
  const spanDuration = spanEvents.pop().span._duration.pop();

  // The '- 1' here is because sometimes the test fails due to the span duration
  // being something like 99.87ms. As long as it's at least 99ms we're happy
  const waitTimeNanos = (waitTimeMillis - 1) * 1000 * 1000;
  expect(spanDuration).toBeGreaterThan(waitTimeNanos);
});

test("withSpan on error", async () => {
  try {
    await tracing.withSpan("name", async () => {
      throw "err";
    });
    // previous line should have an error thrown
    expect(true).toEqual(false);
  } catch (e) {
    expect(e).toEqual("err");
    expect(spanEvents.map((e) => e.event)).toEqual(["onStart", "onEnd"]);
    const lastSpanEvents = spanEvents.pop().span.events;
    expect(lastSpanEvents).length(1);
    expect(lastSpanEvents[0].name).toEqual("exception");
    expect(lastSpanEvents[0].attributes).toEqual({
      "exception.message": "err",
    });
  }
});

test("fetch - 200", async () => {
  const res = await fetch("http://example.com");
  expect(res.status).toEqual(200);

  expect(spanEvents.map((e) => e.event)).toEqual(["onStart", "onEnd"]);
  expect(spanEvents.pop().span.attributes).toEqual({
    "http.url": "http://example.com/",
    "http.scheme": "http",
    "http.method": "GET",
    "http.status": 200,
    "http.status_text": "OK",
  });
});

test("fetch - 404", async () => {
  await fetch("http://example.com/movies.json");

  expect(spanEvents.map((e) => e.event)).toEqual(["onStart", "onEnd"]);
  expect(spanEvents.pop().span.attributes).toEqual({
    "http.url": "http://example.com/movies.json",
    "http.scheme": "http",
    "http.method": "GET",
    "http.status": 404,
    "http.status_text": "Not Found",
  });
});

test("fetch - invalid URL", async () => {
  try {
    await fetch({});
  } catch (err) {
    expect(err.message).toEqual("Invalid URL");
  }

  expect(spanEvents.map((e) => e.event)).toEqual(["onStart", "onEnd"]);

  const span = spanEvents.pop().span;
  expect(spanEvents.pop().span.attributes).toEqual({});
  expect(span.events[0].name).toEqual("exception");
  expect.assertions(4);
});

test("fetch - ENOTFOUND", async () => {
  try {
    await fetch("http://qpwoeuthnvksnvnsanrurvnc.com");
  } catch (err) {
    expect(err.message).toEqual("fetch failed");
    expect(err.cause.code).toEqual("ENOTFOUND");
  }

  expect(spanEvents.map((e) => e.event)).toEqual(["onStart", "onEnd"]);

  const span = spanEvents.pop().span;
  expect(span.attributes).toEqual({
    "http.method": "GET",
    "http.scheme": "http",
    "http.url": "http://qpwoeuthnvksnvnsanrurvnc.com/",
  });
  expect(span.events[0].name).toEqual("exception");
  expect.assertions(5);
});

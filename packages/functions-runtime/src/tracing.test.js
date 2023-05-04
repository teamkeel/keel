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

test("withSpan span time", async () => {
  const waitTimeMillis = 100;
  await tracing.withSpan("name", async () => {
    await new Promise((resolve) => setTimeout(resolve, waitTimeMillis));
  });

  expect(spanEvents.map((e) => e.event)).toEqual(["onStart", "onEnd"]);
  const spanDuration = spanEvents.pop().span._duration.pop();
  const waitTimeNanos = waitTimeMillis * 1000 * 1000;
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

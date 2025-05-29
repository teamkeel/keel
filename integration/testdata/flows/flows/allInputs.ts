import { AllInputs, MyEnum } from "@teamkeel/sdk";

export default AllInputs({}, async (ctx, inputs) => {
  const {
    text,
    number,
    file,
    date,
    timestamp,
    duration,
    bool,
    decimal,
    myEnum,
    markdown,
  } = inputs;

  if (text !== "text") {
    throw new Error("text is not text");
  }

  if (number !== 1) {
    throw new Error("number is not 1");
  }

  const b = await file?.read();
  const contents = await b?.toString("utf-8");

  if (contents !== "hello") {
    throw new Error("file does not contain 'hello'");
  }

  if (date !== new Date("2021-01-01")) {
    throw new Error("date is not 2021-01-01");
  }

  if (timestamp !== new Date("2021-01-01T12:30:15.000Z")) {
    throw new Error("timestamp is not 2021-01-01T12:30:15.000Z");
  }

  if (duration?.toISOString() !== "PT1000S") {
    throw new Error("duration is not PT1000S");
  }

  if (bool !== true) {
    throw new Error("bool is not true");
  }

  if (decimal !== 1.1) {
    throw new Error("decimal is not 1.1");
  }

  if (myEnum !== MyEnum.Value1) {
    throw new Error("enum is not Value1");
  }

  if (markdown !== "**Hello**") {
    throw new Error("markdown is not **Hello**");
  }
});

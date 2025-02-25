import { CreateThing } from "@teamkeel/sdk";

export default CreateThing({
  beforeWrite(ctx, inputs, values) {
    values.texts?.pop();
    values.numbers?.pop();
    values.dates?.pop();
    values.booleans?.pop();
    values.timestamps?.pop();
    values.enums?.pop();
    values.files?.pop();
    values.durations?.pop();

    return {
      ...values,
    };
  },
  afterWrite(ctx, inputs, values) {
    return {
      ...values,
    };
  },
});

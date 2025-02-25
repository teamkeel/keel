import { CreateThingEmpty } from "@teamkeel/sdk";

export default CreateThingEmpty({
  beforeWrite(ctx, inputs, values) {
    if (inputs.files?.[0]?.filename === "one.txt") {
    }

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

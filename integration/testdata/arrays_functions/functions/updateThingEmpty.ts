import { UpdateThingEmpty } from "@teamkeel/sdk";

export default UpdateThingEmpty({
  beforeWrite(ctx, inputs, values) {
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

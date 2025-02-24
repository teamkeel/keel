import { CreateThing } from "@teamkeel/sdk";

export default CreateThing({
  beforeWrite(ctx, inputs, values) {
    return {
      ...values,
    };
  },
});

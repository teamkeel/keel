import { CreateUser } from "@teamkeel/sdk";

export default CreateUser({
  beforeWrite(ctx, inputs, values) {
    return {
      ...values,
      identityId: ctx.identity!.id,
    };
  },
});

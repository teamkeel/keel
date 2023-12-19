import { models, CreateAdmission } from "@teamkeel/sdk";

export default CreateAdmission({
  async beforeWrite(ctx, inputs, values) {
    const audience = await models.audience.findOne({
      identityId: ctx.identity!.id,
    });

    return {
      ...values,
      audienceId: audience!.id,
    };
  },
});

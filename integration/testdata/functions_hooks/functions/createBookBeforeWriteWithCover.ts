import {
  CreateBookBeforeWriteWithCover,
  CreateBookBeforeWriteWithCoverHooks,
} from "@teamkeel/sdk";

export default CreateBookBeforeWriteWithCover({
  async beforeWrite(ctx, inputs, values) {
    return {
      ...values,
      title: values.title.toUpperCase(),
    };
  },
});

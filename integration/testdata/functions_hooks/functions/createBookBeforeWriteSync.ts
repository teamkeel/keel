import { CreateBookBeforeWriteSync } from "@teamkeel/sdk";

export default CreateBookBeforeWriteSync({
  // Note: this is testing a sync hook so very important there is no "async" keyword before the function
  beforeWrite(ctx, inputs, values) {
    return {
      ...values,
      title: values.title.toUpperCase(),
    };
  },
});

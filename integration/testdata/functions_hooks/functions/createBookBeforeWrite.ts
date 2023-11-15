import { CreateBookBeforeWrite } from "@teamkeel/sdk";

// This functon tests that a beforeWrite hook in a create function
// can return new values to be used for creating the record, based
// on the provided values
export default CreateBookBeforeWrite({
  async beforeWrite(ctx, inputs, values) {
    return {
      ...values,
      title: values.title.toUpperCase(),
    };
  },
});

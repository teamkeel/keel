import { CreateBookWithAuthor } from "@teamkeel/sdk";

export default CreateBookWithAuthor({
  beforeWrite(ctx, inputs, values) {
    // assert input format
    if (typeof inputs.author!.id !== "string") {
      throw new Error("expected inputs.author.id to be a string");
    }

    // assert values format
    if (typeof values.author!.id !== "string") {
      throw new Error("expected values.author.id to be a string");
    }

    return values;
  },
});

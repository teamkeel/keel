import { CreateBookAndAuthor } from "@teamkeel/sdk";

export default CreateBookAndAuthor({
  beforeWrite(ctx, inputs, values) {
    // assert input format
    if (typeof inputs.author!.name !== "string") {
      throw new Error("expected inputs.author.name to be a string");
    }

    // assert values format
    if (typeof values.author!.name !== "string") {
      throw new Error("expected values.author.name to be a string");
    }

    return values;
  },
});

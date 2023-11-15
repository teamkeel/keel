import { CreateBookAfterWriteErrorRollback } from "@teamkeel/sdk";

export default CreateBookAfterWriteErrorRollback({
  afterWrite(ctx, inputs, data) {
    if (data.title.toLowerCase() == "lady chatterley's lover") {
      // Throwing an error should cause the transaction to be rolled back
      throw new Error("this book is banned");
    }
  },
});

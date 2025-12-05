import { CreateBookNoTransaction, models } from "@teamkeel/sdk";

// Test: dbTransaction: false disables automatic transaction wrapping
// When disabled, records created before an error are NOT rolled back
const fn = CreateBookNoTransaction({
  async afterWrite(ctx, inputs, data) {
    // Create an additional record that would normally be rolled back
    await models.deletedBook.create({
      bookId: data.id,
      title: "no-transaction-test-" + data.title,
      deletedAt: ctx.now(),
    });

    // Throw an error - without transaction, the deletedBook record should persist
    throw new Error("error after creating deletedBook");
  },
}) as any;

// Disable automatic transaction wrapping
fn.config = { dbTransaction: false };

export default fn;

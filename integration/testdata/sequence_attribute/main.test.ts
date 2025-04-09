import { actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("basic actions", async () => {
  const first = await actions.createInvoice({
    amount: 100,
  });
  expect(first.reference).toBe("INV-0001");
  expect("referenceSequence" in first).toBe(false);

  const second = await actions.createInvoice({
    amount: 200,
  });
  // assert that the sequence is incrementing
  expect(second.reference).toBe("INV-0002");

  // should be able to fetch by the reference value
  let getResponse = await actions.getInvoice({ reference: first.reference });
  expect(getResponse?.id).toBe(first.id);

  // should be able to update by the reference value
  const updateResponse = await actions.updateInvoice({
    where: {
      reference: first.reference,
    },
    values: {
      paid: true,
    },
  });
  expect(updateResponse?.id).toBe(first.id);
  expect(updateResponse?.paid).toBe(true);

  // should be able to delete by reference
  const deleteResponse = await actions.deleteInvoice({
    reference: first.reference,
  });
  expect(deleteResponse).toBe(first.id);

  // invoice should have been deleted
  getResponse = await actions.getInvoice({ reference: first.reference });
  expect(getResponse).toBe(null);
});

test("functions", async () => {
  const first = await actions.createInvoiceFunc({
    amount: 100,
  });
  expect(first.reference).toBe("ord_0001");
  expect("referenceSequence" in first).toBe(false);

  const second = await actions.createInvoiceFunc({
    amount: 200,
  });
  // assert that the sequence is incrementing
  expect(second.reference).toBe("ord_0002");

  // should be able to fetch by the reference value
  let getResponse = await actions.getInvoiceFunc({
    reference: first.reference,
  });
  expect(getResponse?.id).toBe(first.id);

  // should be able to update by the reference value
  const updateResponse = await actions.updateInvoiceFunc({
    where: {
      reference: first.reference,
    },
    values: {
      paid: true,
    },
  });
  expect(updateResponse?.id).toBe(first.id);
  expect(updateResponse?.paid).toBe(true);

  // should be able to delete by reference
  const deleteResponse = await actions.deleteInvoiceFunc({
    reference: first.reference,
  });
  expect(deleteResponse).toBe(first.id);

  // invoice should have been deleted
  getResponse = await actions.getInvoiceFunc({ reference: first.reference });
  expect(getResponse).toBe(null);
});

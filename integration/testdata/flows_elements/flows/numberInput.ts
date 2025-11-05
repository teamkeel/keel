import { NumberInput, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default NumberInput(config, async (ctx) => {
  const page1 = await ctx.ui.page("number input page", {
    content: [
      ctx.ui.inputs.number("age", {
        label: "Age",
        defaultValue: 25,
        validate: (data) => {
          // Age must be between 0 and 150
          if (data < 0) {
            return "Age cannot be negative";
          }
          if (data > 150) {
            return "Age must be 150 or less";
          }
          return true;
        },
      }),
      ctx.ui.inputs.number("quantity", {
        label: "Quantity",
        optional: true,
        validate: (data) => {
          // Only validate if quantity is provided (since it's optional)
          if (data !== null && data !== undefined && data < 1) {
            return "Quantity must be at least 1";
          }
          return true;
        },
      }),
      ctx.ui.inputs.number("price", {
        label: "Price",
        helpText: "Enter the price in dollars",
        defaultValue: 0,
        validate: (data) => {
          // Price must be non-negative
          if (data < 0) {
            return "Price cannot be negative";
          }
          return true;
        },
      }),
    ],
  });

  return {
    age: page1.age,
    quantity: page1.quantity,
    price: page1.price,
  };
});

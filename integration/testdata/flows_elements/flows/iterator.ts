import { Iterator, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default Iterator(config, async (ctx) => {
  await ctx.ui.page("my page", {
    content: [
      ctx.ui.iterator("my iterator", {
        content: [
          ctx.ui.display.header({
            level: 1,
            title: "my header",
            description: "my description",
          }),
          ctx.ui.select.one("sku", {
            label: "SKU",
            options: [
              "SHOES",
              "SHIRTS",
              "PANTS",
              "TIE",
              "BELT",
              "SOCKS",
              "UNDERWEAR",
            ],
            optional: false,
          }),
          ctx.ui.inputs.number("quantity", {
            label: "Qty",
            optional: false,
            validate: (value) => {
              if (value < 1) {
                return "Quantity must be greater than 0";
              } else if (value > 10) {
                return "Quantity must be less than 10";
              }
              return true;
            },
          }),
        ],
        min: 1,
      }),
    ],
    validate: (data) => {
      let totalQuantity = 0;
      for (const item of data["my iterator"]) {
        totalQuantity += item.quantity;
      }
      if (totalQuantity > 20) {
        return "Total quantity must be less than 20";
      }
      return true;
    },
  });

  return null;
});

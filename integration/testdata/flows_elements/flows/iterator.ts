import { Iterator, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default Iterator(config, async (ctx) => {
  const result = await ctx.ui.page("my page", {
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
          }),
          ctx.ui.inputs.number("quantity", {
            label: "Qty",
          }),
        ],
        min: 1,
        max: 5,
      }),
    ],
  });

  return { data: result };
});

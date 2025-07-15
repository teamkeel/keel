import { ValidationText } from "@teamkeel/sdk";

const postcodeRegex = /^[A-Za-z]{1,2}\d[A-Za-z\d]?\s*\d[A-Za-z]{2}$/;

export default ValidationText({}, async (ctx) => {
  await ctx.ui.page("my page", {
    title: "Your postcode",
    content: [
      ctx.ui.inputs.text("postcode", {
        label: "Postcode",
        placeholder: "e.g. N1 ABC",
        validate(value) {
          if (postcodeRegex.test(value)) {
            return true;
          }
          return "not a valid postcode";
        },
      }),
    ],
  });
});

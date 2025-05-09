import { ValidationBoolean } from "@teamkeel/sdk";

export default ValidationBoolean({}, async (ctx) => {
  await ctx.ui.page("first page", {
    title: "Important question",
    content: [
      ctx.ui.inputs.boolean("good", {
        label: "Is it good?",
        // validate can be an async function
        async validate(value) {
          if (!value) {
            return "it must be good";
          }

          return null;
        },
      }),
    ],
  });
});

import { ErrorInValidation } from "@teamkeel/sdk";

const emailRegex = /^[^@]+@[^@]+\.[^@]+$/;
const phoneRegex = /^[0-9]{10}$/;

export default ErrorInValidation({}, async (ctx) => {
  await ctx.ui.page("first page", {
    content: [
      ctx.ui.inputs.text("email", {
        label: "Email",
        optional: true,
      }),
    ],
    validate(value) {
      throw new Error("something has gone wrong");
    },
  });
});

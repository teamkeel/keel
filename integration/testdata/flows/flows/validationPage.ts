import { ValidationPage } from "@teamkeel/sdk";

const emailRegex = /^[^@]+@[^@]+\.[^@]+$/;
const phoneRegex = /^[0-9]{10}$/;

export default ValidationPage({}, async (ctx) => {
  await ctx.ui.page("first page", {
    content: [
      ctx.ui.inputs.text("email", {
        label: "Email",
        optional: true,
        validate(value) {
          if (!emailRegex.test(value)) {
            return "Not a valid email";
          }
          return true;
        },
      }),
      ctx.ui.inputs.text("phone", {
        label: "Phone",
        optional: true,
        validate(value) {
          if (!phoneRegex.test(value)) {
            return "Not a valid phone number";
          }
          return true;
        },
      }),
    ],
    validate: async (data) => {
      if (!data.email && !data.phone) {
        return "Email or phone is required";
      }
      return true;
    },
  });
});

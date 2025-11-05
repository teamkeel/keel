import { DatePickerWithActions, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default DatePickerWithActions(config, async (ctx) => {
  const page1 = await ctx.ui.page("date picker validation", {
    content: [
      ctx.ui.inputs.datePicker("startDate", {
        label: "Start Date",
        validate: (data, action) => {
          // Verify action parameter is passed correctly when an action is provided
          if (
            action !== undefined &&
            action !== "schedule" &&
            action !== "draft"
          ) {
            throw new Error(
              `Expected action to be 'schedule' or 'draft', got: ${action}`
            );
          }

          if (action === "schedule") {
            // When scheduling, date must be in the future
            const startDate = new Date(data);
            const today = new Date();
            today.setHours(0, 0, 0, 0);
            if (startDate < today) {
              return "Start date must be in the future for scheduling";
            }
          }
          // Draft allows any date
          return true;
        },
      }),
    ],
    actions: ["schedule", "draft"],
  });

  return {
    action: page1.action,
    startDate: page1.data.startDate,
  };
});

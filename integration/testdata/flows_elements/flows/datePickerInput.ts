import { DatePickerInput, FlowConfig } from "@teamkeel/sdk";

const config = {
  // See https://docs.keel.so/flows for options
} as const satisfies FlowConfig;

export default DatePickerInput(config, async (ctx) => {
  const defaultDate = "2024-01-01";

  const page1 = await ctx.ui.page("date picker page", {
    content: [
      ctx.ui.inputs.datePicker("birthDate", {
        label: "Birth Date",
        defaultValue: defaultDate,
        validate: (data) => {
          // Birth date must be in the past
          const birthDate = new Date(data);
          const now = new Date();
          now.setHours(0, 0, 0, 0);
          if (birthDate >= now) {
            return "Birth date must be in the past";
          }
          // Birth date must be after 1900
          const minDate = new Date("1900-01-01");
          if (birthDate < minDate) {
            return "Birth date must be after 1900";
          }
          return true;
        },
      }),
      ctx.ui.inputs.datePicker("startDate", {
        label: "Start Date",
        optional: true,
        validate: (data) => {
          // If provided, start date must not be more than 1 year in the future
          if (data) {
            const startDate = new Date(data);
            const oneYearFromNow = new Date();
            oneYearFromNow.setFullYear(oneYearFromNow.getFullYear() + 1);
            if (startDate > oneYearFromNow) {
              return "Start date cannot be more than 1 year in the future";
            }
          }
          return true;
        },
      }),
      ctx.ui.inputs.datePicker("appointmentDate", {
        label: "Appointment Date",
        helpText: "Select your preferred appointment date",
        validate: (data) => {
          // Appointment must be in the future
          const appointmentDate = new Date(data);
          const today = new Date();
          today.setHours(0, 0, 0, 0);
          if (appointmentDate < today) {
            return "Appointment date must be in the future";
          }
          return true;
        },
      }),
    ],
  });

  return {
    birthDate: page1.birthDate,
    startDate: page1.startDate,
    appointmentDate: page1.appointmentDate,
  };
});

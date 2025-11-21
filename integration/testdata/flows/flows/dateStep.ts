import { DateStep, models } from "@teamkeel/sdk";

export default DateStep(
    {
        title: "Date step",
    },
    async (ctx) => {
        const dateResult = await ctx.step("create date", async () => {
            const now = new Date("2024-01-15T10:30:00.000Z");
            return now;
        });

        // Verify the returned value is actually a Date object
        await ctx.step("verify date object", async () => {
            // Cast to Date since we know it will be deserialized as a Date
            const date = dateResult as Date;

            return {
                isDate: date instanceof Date,
                isoString: date.toISOString(),
                timestamp: date.getTime(),
            };
        });
    }
);


import { WriteThing, models } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default WriteThing(async (ctx, inputs) => {
  const thing = await models.thing.create({
    texts: inputs.texts,
    numbers: inputs.numbers,
    dates: inputs.dates,
    booleans: inputs.booleans,
    timestamps: inputs.timestamps,
    enums: inputs.enums,
    files: inputs.files,
    durations: inputs.durations,
  });

  return { thing };
});

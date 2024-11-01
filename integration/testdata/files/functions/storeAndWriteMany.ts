import { StoreAndWriteMany, models } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default StoreAndWriteMany(async (ctx, inputs) => {
  const stored = await inputs.file.store();

  await models.myFile.create({ file: stored });
  await models.myFile.create({ file: stored });
  await models.myFile.create({ file: stored });

  return { msg: { file: stored } };
});

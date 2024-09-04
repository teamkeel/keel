import { WriteMany, models } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default WriteMany(async (ctx, inputs) => {
  await models.myFile.create({ file: inputs.msg.file });
  await models.myFile.create({ file: inputs.msg.file });
  await models.myFile.create({ file: inputs.msg.file });

  return { msg: { file: inputs.msg.file } };
});

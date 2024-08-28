import { WriteMany, InlineFile, StoredFile, models } from '@teamkeel/sdk';

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default WriteMany(async (ctx, inputs) => {

    const f1 = await models.myFile.create({ file: inputs.file });
    const f2 = await models.myFile.create({ file: inputs.file });
    const f3 = await models.myFile.create({ file: inputs.file });

    return "";

});
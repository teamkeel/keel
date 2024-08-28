import { StoreAndWriteMany, InlineFile, StoredFile, models } from '@teamkeel/sdk';

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default StoreAndWriteMany(async (ctx, inputs) => {
    const stored = await inputs.file.store();

    const f1 = await models.myFile.create({ file: stored });
    const f2 = await models.myFile.create({ file: stored });
    const f3 = await models.myFile.create({ file: stored });

    return "";
});
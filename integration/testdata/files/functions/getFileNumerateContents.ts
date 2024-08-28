import { GetFileNumerateContents, GetFileNumerateContentsHooks, models } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks : GetFileNumerateContentsHooks = {};

export default GetFileNumerateContents({
    beforeQuery: async (ctx, inputs, query) => {
        const myFile = await models.myFile.findOne({id:inputs.id});

        console.log(myFile?.file);

        const buffer = await myFile?.file?.read();
        const contents = buffer?.toString("utf-8");

        const number = parseInt(contents!, 10);

        const next = (number + 1).toString();
        const buffer2 = Buffer.from(next)
         myFile?.file?.write(buffer2);
        await myFile?.file?.store();
        
        return query;
      },
});

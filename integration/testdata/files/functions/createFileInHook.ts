import { CreateFileInHook, CreateFileInHookHooks, InlineFile } from '@teamkeel/sdk';

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks : CreateFileInHookHooks = {};

export default CreateFileInHook({
    beforeWrite: async (ctx, inputs) => {
        const fileContents = "created in hook!";
        const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from( fileContents).toString("base64")}`;
        const file = InlineFile.fromDataURL(dataUrl);

        return {
          file: file
        };
      },
});
	
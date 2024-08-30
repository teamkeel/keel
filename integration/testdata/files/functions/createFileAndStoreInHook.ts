import {
  CreateFileAndStoreInHook,
  CreateFileAndStoreInHookHooks,
  InlineFile,
  StoredFile,
} from "@teamkeel/sdk";

// To learn more about what you can do with hooks, visit https://docs.keel.so/functions
const hooks: CreateFileAndStoreInHookHooks = {};

export default CreateFileAndStoreInHook({
  beforeWrite: async (ctx, inputs) => {
    const fileContents = "created and stored in hook!";
    const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
      fileContents
    ).toString("base64")}`;

    const file = InlineFile.fromDataURL(dataUrl);
    const storedFile = await file.store();

    return {
      file: storedFile,
    };
  },
});

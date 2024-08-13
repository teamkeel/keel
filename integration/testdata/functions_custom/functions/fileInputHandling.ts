import { FileInputHandling, permissions } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default FileInputHandling(async (ctx, inputs) => {
  permissions.allow();

  const fileData = await inputs.file.store();

  return {
    filename: fileData.filename,
    size: fileData.size,
    contentType: fileData.contentType,
    key: fileData.key,
  };
});

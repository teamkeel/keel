import { FileInputHandling, permissions } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default FileInputHandling(async (ctx, inputs) => {
  permissions.allow();

  return {
    filename: inputs.file.filename,
    size: inputs.file.size,
    contentType: inputs.file.contentType,
  };
});

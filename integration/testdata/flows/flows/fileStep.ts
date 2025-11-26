import { FileStep, models } from "@teamkeel/sdk";
import { InlineFile } from "@teamkeel/functions-runtime";

export default FileStep(
  {
    title: "File step",
  },
  async (ctx) => {
    const fileResult = await ctx.step("create and store file", async () => {
      // Create a file from a data URL
      const inlineFile = InlineFile.fromDataURL(
        "data:text/plain;name=test-file.txt;base64,SGVsbG8gV29ybGQh"
      );
      
      // Store the file to get a File instance with a key
      const storedFile = await inlineFile.store();
      
      // Return the File instance
      return storedFile as any;
    });

    // Verify the returned File can be used in subsequent steps
    await ctx.step("verify file object", async () => {
      // Cast to any since the type system doesn't know it's a File
      const file = fileResult as any;
      
      // Read the file contents
      const buffer = await file.read();
      const contents = buffer.toString("utf-8");
      
      return {
        hasFilename: file.filename === "test-file.txt",
        hasContentType: file.contentType === "text/plain",
        hasKey: typeof file._key === "string" && file._key.length > 0,
        hasSize: file.size > 0,
        canRead: contents === "Hello World!",
        isFileInstance: file.constructor.name === "File",
      };
    });
  }
);


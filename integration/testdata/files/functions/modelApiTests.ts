import { ModelApiTests, InlineFile, models } from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default ModelApiTests(async (ctx, inputs) => {
  const file = new InlineFile({
    filename: "file.csv",
    contentType: "text/csv",
  });

  file.write(Buffer.from("hello world"));
  const b1 = await file.read();
  if (b1.toString("utf-8") != "hello world") {
    throw new Error("reading a locally instantiated file failed");
  }

  const createdFile = await models.myFile.create({ file: file });
  const b2 = await createdFile.file?.read();
  if (b2?.toString("utf-8") != "hello world") {
    throw new Error("reading a file created using models.created failed");
  }

  const findFile = await models.myFile.findOne({ id: createdFile.id });
  const b3 = await findFile?.file?.read();
  if (b3?.toString("utf-8") != "hello world") {
    throw new Error("reading a file retrieived using models.findOne failed");
  }

  const whereFiles = await models.myFile.findMany();
  const b9 = await whereFiles[0]?.file?.read();
  if (b9?.toString("utf-8") != "hello world") {
    throw new Error("reading a file retrieived using models.findMany failed");
  }

  const whereFilesQB = await models.myFile
    .where({ id: createdFile.id })
    .findMany();
  const b4 = await whereFilesQB[0]?.file?.read();
  if (b4?.toString("utf-8") != "hello world") {
    throw new Error(
      "reading a file retrieived using models.where.findMany (the query builder) failed"
    );
  }

  file.write(Buffer.from("goodbye world"));
  const stored = await file.store();
  const b5 = await stored.read();
  if (b5.toString("utf-8") != "goodbye world") {
    throw new Error("reading a locally instantiated and stored file failed");
  }

  const updatedFile = await models.myFile.update(
    { id: createdFile.id },
    { file: stored }
  );
  const b6 = await updatedFile.file?.read();
  if (b6?.toString("utf-8") != "goodbye world") {
    throw new Error(
      "reading a stored file updated using models.updated failed"
    );
  }

  stored.write(Buffer.from("hello again!"));
  await stored.store();
  const b7 = await stored.read();
  if (b7.toString("utf-8") != "hello again!") {
    throw new Error("reading a restored file failed");
  }

  const findFile2 = await models.myFile.findOne({ id: createdFile.id });
  const b8 = await findFile2?.file?.read();
  if (b8?.toString("utf-8") != "hello again!") {
    throw new Error("reading a file which was rewritten to failed");
  }

  file.write(Buffer.from("goodbye again world"));
  const b11 = await file.read();
  if (b11.toString("utf-8") != "goodbye again world") {
    throw new Error("reading a locally instantiated and stored file failed");
  }

  const updatedFile2 = await models.myFile.update(
    { id: createdFile.id },
    { file: file! }
  );
  const b10 = await updatedFile2.file?.read();
  if (b10?.toString("utf-8") != "goodbye again world") {
    throw new Error(
      "reading a inline file updated using models.updated failed"
    );
  }

  return "";
});

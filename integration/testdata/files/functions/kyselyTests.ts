import {
  KyselyTests,
  InlineFile,
  File,
  models,
  useDatabase,
} from "@teamkeel/sdk";

// To learn more about what you can do with custom functions, visit https://docs.keel.so/functions
export default KyselyTests(async (ctx, inputs) => {
  const file = new InlineFile({
    filename: "hi.txt",
    contentType: "text/plain",
  });

  file.write(Buffer.from("hello world"));

  const stored = await file.store();

  const row = await useDatabase()
    .insertInto("my_file")
    .values({
      file: stored.toDbRecord(),
    })
    .returningAll()
    .executeTakeFirst();

  if (row?.file?.key != stored.key) {
    throw new Error("stored file with kysely not matching file");
  }

  const record = await models.myFile.findOne({ id: row.id });

  const buffer = await record?.file?.read();
  if (buffer?.toString("utf-8") != "hello world") {
    throw new Error("reading a file stored with kysely failed");
  }

  return "";
});

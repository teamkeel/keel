import { actions, resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";
import { useDatabase, InlineFile } from "@teamkeel/sdk";
import { sql } from "kysely";

interface DbFile {
  id: string;
  data: any;
  filename: string;
  contentType: string;
  createdAt: Date;
}

beforeEach(resetDatabase);

test("files - create action with file input", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  expect(result.file?.contentType).toEqual("application/text");
  expect(result.file?.filename).toEqual("my-file.txt");
  expect(result.file?.size).toEqual(5);

  const contents1 = await result.file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello");

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files = await sql<DbFile>`SELECT * FROM keel_storage`.execute(
    useDatabase()
  );

  expect(myfiles.length).toEqual(1);
  expect(files.rows.length).toEqual(1);
  expect(files.rows[0].id).toEqual(myfiles[0].file?.key);
  expect(files.rows[0].filename).toEqual(myfiles[0].file?.filename);
  expect(files.rows[0].contentType).toEqual(myfiles[0].file?.contentType);

  const contents = files.rows[0].data.toString("utf-8");
  expect(contents).toEqual("hello");
});

test("files - update action with file input", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const fileContents2 = "hello again";
  const dataUrl2 = `data:application/text;name=my-second-file.txt;base64,${Buffer.from(
    fileContents2
  ).toString("base64")}`;

  const updated = await actions.updateFile({
    where: {
      id: result.id,
    },
    values: {
      file: InlineFile.fromDataURL(dataUrl2),
    },
  });

  expect(updated.file?.contentType).toEqual("application/text");
  expect(updated.file?.filename).toEqual("my-second-file.txt");
  expect(updated.file?.size).toEqual(11);

  const contents1 = await updated.file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello again");

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<DbFile>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
      useDatabase()
    );

  expect(myfiles.length).toEqual(1);
  expect(files.rows.length).toEqual(2);
  expect(files.rows[0].id).toEqual(myfiles[0].file?.key);
  expect(files.rows[0].filename).toEqual(myfiles[0].file?.filename);
  expect(files.rows[0].contentType).toEqual(myfiles[0].file?.contentType);

  const contents = files.rows[0].data.toString("utf-8");
  expect(contents).toEqual("hello again");
});

test("files - get action", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const created = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const result = await actions.getFile({
    id: created.id,
  });

  expect(result?.file?.contentType).toEqual("application/text");
  expect(result?.file?.filename).toEqual("my-file.txt");
  expect(result?.file?.size).toEqual(5);

  const contents1 = await result?.file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello");
});

test("files - list action", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const created1 = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const fileContents2 = "hello again";
  const dataUrl2 = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents2
  ).toString("base64")}`;

  const created2 = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl2),
  });

  const result = await actions.listFiles({});

  expect(result.results[0].file?.contentType).toEqual("application/text");
  expect(result.results[0].file?.filename).toEqual("my-file.txt");
  expect(result.results[0].file?.size).toEqual(5);

  const contents1 = await result.results[0].file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello");

  expect(result.results[1].file?.contentType).toEqual("application/text");
  expect(result.results[1].file?.filename).toEqual("my-file.txt");
  expect(result.results[1].file?.size).toEqual(11);

  const contents2 = await result.results[1].file?.read();
  expect(contents2?.toString("utf-8")).toEqual("hello again");
});

test("files - create file in hook", async () => {
  await actions.createFileInHook({});

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<DbFile>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
      useDatabase()
    );

  expect(myfiles.length).toEqual(1);
  expect(files.rows.length).toEqual(1);
  expect(files.rows[0].id).toEqual(myfiles[0].file?.key);
  expect(files.rows[0].filename).toEqual(myfiles[0].file?.filename);
  expect(files.rows[0].contentType).toEqual(myfiles[0].file?.contentType);

  const contents = files.rows[0].data.toString("utf-8");
  expect(contents).toEqual("created in hook!");
});

test("files - create and store file in hook", async () => {
  await actions.createFileAndStoreInHook({});

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<DbFile>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
      useDatabase()
    );

  expect(myfiles.length).toEqual(1);
  expect(files.rows.length).toEqual(1);
  expect(files.rows[0].id).toEqual(myfiles[0].file?.key);
  expect(files.rows[0].filename).toEqual(myfiles[0].file?.filename);
  expect(files.rows[0].contentType).toEqual(myfiles[0].file?.contentType);

  const contents = files.rows[0].data.toString("utf-8");
  expect(contents).toEqual("created and stored in hook!");
});

test("files - read and store in query hook", async () => {
  const fileContents = "1";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  await actions.getFileNumerateContents({ id: result.id });
  await actions.getFileNumerateContents({ id: result.id });
  await actions.getFileNumerateContents({ id: result.id });

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<DbFile>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
      useDatabase()
    );

  expect(myfiles.length).toEqual(1);
  expect(files.rows.length).toEqual(1);
  expect(files.rows[0].id).toEqual(myfiles[0].file?.key);
  expect(files.rows[0].filename).toEqual(myfiles[0].file?.filename);
  expect(files.rows[0].contentType).toEqual(myfiles[0].file?.contentType);

  const contents = files.rows[0].data.toString("utf-8");
  expect(contents).toEqual("4");
});

test("files - write many, store many", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.writeMany({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<DbFile>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
      useDatabase()
    );

  expect(myfiles.length).toEqual(3);
  expect(files.rows.length).toEqual(3);
  expect(myfiles[0].file?.key).toEqual(files.rows[0].id);
  expect(myfiles[1].file?.key).toEqual(files.rows[1].id);
  expect(myfiles[2].file?.key).toEqual(files.rows[2].id);
});

test("files - store once, write many", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.storeAndWriteMany({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<DbFile>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
      useDatabase()
    );

  expect(myfiles.length).toEqual(3);
  expect(files.rows.length).toEqual(1);
  expect(myfiles[0].file?.key).toEqual(files.rows[0].id);
  expect(myfiles[1].file?.key).toEqual(files.rows[0].id);
  expect(myfiles[2].file?.key).toEqual(files.rows[0].id);
});

test("files - model API file tests", async () => {
  await expect(actions.modelApiTests({})).not.toHaveError({});
});

test("files - kysely file tests", async () => {
  await expect(actions.kyselyTests({})).not.toHaveError({});
});

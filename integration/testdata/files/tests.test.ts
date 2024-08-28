import { actions, resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";
import { useDatabase } from "@teamkeel/sdk";
import { sql } from "kysely";

interface File {
  id: string;
  data: any;
  filename: string;
  contentType: string;
  createdAt: Date;
}

beforeEach(resetDatabase);

test("create action with file input", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  await actions.createFile({
    file: dataUrl,
  });

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files = await sql<File>`SELECT * FROM keel_storage`.execute(
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

test("update action with file input", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.createFile({
    file: dataUrl,
  });

  const fileContents2 = "hello again";
  const dataUrl2 = `data:application/text;name=my-second-file.txt;base64,${Buffer.from(
    fileContents2
  ).toString("base64")}`;

  await actions.updateFile({
    where: {
      id: result.id,
    },
    values: {
      file: dataUrl2,
    },
  });

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<File>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
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

test("create file in hook", async () => {
  await actions.createFileInHook({});

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<File>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
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

test("create and store file in hook", async () => {
  await actions.createFileAndStoreInHook({});

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<File>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
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

test("read and store in query hook", async () => {
  const fileContents = "1";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.createFile({
    file: dataUrl,
  });

  await actions.getFileNumerateContents({ id: result.id });
  await actions.getFileNumerateContents({ id: result.id });
  await actions.getFileNumerateContents({ id: result.id });

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<File>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
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

test("write many, store many", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.writeMany({
    file: dataUrl,
  });

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<File>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
      useDatabase()
    );

  expect(myfiles.length).toEqual(3);
  expect(files.rows.length).toEqual(3);
  expect(myfiles[0].file?.key).toEqual(files.rows[0].id);
  expect(myfiles[1].file?.key).toEqual(files.rows[1].id);
  expect(myfiles[2].file?.key).toEqual(files.rows[2].id);
});

test("store once, write many", async () => {
  const fileContents = "hello";
  const dataUrl = `data:application/text;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.storeAndWriteMany({
    file: dataUrl,
  });

  const myfiles = await useDatabase()
    .selectFrom("my_file")
    .selectAll()
    .execute();

  const files =
    await sql<File>`SELECT * FROM keel_storage ORDER BY created_at DESC`.execute(
      useDatabase()
    );

  expect(myfiles.length).toEqual(3);
  expect(files.rows.length).toEqual(1);
  expect(myfiles[0].file?.key).toEqual(files.rows[0].id);
  expect(myfiles[1].file?.key).toEqual(files.rows[0].id);
  expect(myfiles[2].file?.key).toEqual(files.rows[0].id);
});

test("model API file tests", async () => {
  await expect(actions.modelApiTests({})).not.toHaveError({});
});

test("Kysely file tests", async () => {
  await expect(actions.kyselyTests({})).not.toHaveError({});
});

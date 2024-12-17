import { actions, resetDatabase, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";
import { InlineFile, File } from "@teamkeel/sdk";

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
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  expect(result.file?.contentType).toEqual("text/plain");
  expect(result.file?.filename).toEqual("my-file.txt");
  expect(result.file?.size).toEqual(5);

  const contents1 = await result.file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello");

  const myfile = await models.myFile.findOne({ id: result.id });

  expect(myfile?.file?.contentType).toEqual("text/plain");
  expect(myfile?.file?.filename).toEqual("my-file.txt");
});

test("files - update action with file input", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const fileContents2 = "hello again";
  const dataUrl2 = `data:text/plain;name=my-second-file.txt;base64,${Buffer.from(
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

  expect(updated.file?.contentType).toEqual("text/plain");
  expect(updated.file?.filename).toEqual("my-second-file.txt");
  expect(updated.file?.size).toEqual(11);

  const contents1 = await updated.file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello again");

  const myfiles = await models.myFile.findMany();
  expect(myfiles.length).toEqual(1);

  const contents = (await myfiles[0].file?.read())?.toString("utf-8");
  expect(contents).toEqual("hello again");
});

test("files - update action with file input and empty hooks", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const fileContents2 = "hello again";
  const dataUrl2 = `data:text/plain;name=my-second-file.txt;base64,${Buffer.from(
    fileContents2
  ).toString("base64")}`;

  const updated = await actions.updateFileEmptyHooks({
    where: {
      id: result.id,
    },
    values: {
      file: InlineFile.fromDataURL(dataUrl2),
    },
  });

  expect(updated.file?.contentType).toEqual("text/plain");
  expect(updated.file?.filename).toEqual("my-second-file.txt");
  expect(updated.file?.size).toEqual(11);

  // key should have changed
  expect(updated.file?.key).not.toEqual(result.file?.key);

  const contents1 = await updated.file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello again");

  const myfiles = await models.myFile.findMany();
  expect(myfiles.length).toEqual(1);
  expect(myfiles[0].id).toEqual(updated.id);
  expect(myfiles[0].file?.filename).toEqual("my-second-file.txt");

  const contents = (await myfiles[0].file?.read())?.toString("utf-8");
  expect(contents).toEqual("hello again");
});

test("files - get action", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const created = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const result = await actions.getFile({
    id: created.id,
  });

  expect(result?.file?.contentType).toEqual("text/plain");
  expect(result?.file?.filename).toEqual("my-file.txt");
  expect(result?.file?.size).toEqual(5);

  const contents1 = await result?.file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello");
});

test("files - list action", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const fileContents2 = "hello again";
  const dataUrl2 = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents2
  ).toString("base64")}`;

  await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl2),
  });

  const result = await actions.listFiles({});

  expect(result.results[0].file?.contentType).toEqual("text/plain");
  expect(result.results[0].file?.filename).toEqual("my-file.txt");
  expect(result.results[0].file?.size).toEqual(5);

  const contents1 = await result.results[0].file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello");

  expect(result.results[1].file?.contentType).toEqual("text/plain");
  expect(result.results[1].file?.filename).toEqual("my-file.txt");
  expect(result.results[1].file?.size).toEqual(11);

  const contents2 = await result.results[1].file?.read();
  expect(contents2?.toString("utf-8")).toEqual("hello again");
});

test("files - get action empty hooks", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const created = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const result = await actions.getFileEmptyHooks({
    id: created.id,
  });

  expect(result?.file?.contentType).toEqual("text/plain");
  expect(result?.file?.filename).toEqual("my-file.txt");
  expect(result?.file?.size).toEqual(5);

  const contents1 = await result?.file?.read();
  expect(contents1?.toString("utf-8")).toEqual("hello");
});

test("files - list action empty hooks", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  const result = await actions.listFilesEmptyHooks({});

  expect(result.results[0].file?.contentType).toEqual("text/plain");
  expect(result.results[0].file?.filename).toEqual("my-file.txt");
  expect(result.results[0].file?.size).toEqual(5);

  const contents = await result.results[0].file?.read();
  expect(contents?.toString("utf-8")).toEqual("hello");
});

test("files - create file in hook", async () => {
  await actions.createFileInHook({});

  const myfiles = await models.myFile.findMany();
  expect(myfiles.length).toEqual(1);

  const contents = (await myfiles[0].file?.read())?.toString("utf-8");
  expect(contents).toEqual("created in hook!");
});

test("files - create and store file in hook", async () => {
  await actions.createFileAndStoreInHook({});

  const myfiles = await models.myFile.findMany();
  expect(myfiles.length).toEqual(1);

  const contents = (await myfiles[0].file?.read())?.toString("utf-8");
  expect(contents).toEqual("created and stored in hook!");
});

test("files - read and store in query hook", async () => {
  const fileContents = "1";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.createFile({
    file: InlineFile.fromDataURL(dataUrl),
  });

  await actions.getFileNumerateContents({ id: result.id });
  await actions.getFileNumerateContents({ id: result.id });
  const res = await actions.getFileNumerateContents({ id: result.id });

  const myfiles = await models.myFile.findMany();
  expect(myfiles.length).toEqual(1);

  const contents = (await myfiles[0].file?.read())?.toString("utf-8");
  expect(contents).toEqual("4");
});

test("files - write many, store many", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.writeMany({
    msg: { file: InlineFile.fromDataURL(dataUrl) },
  });

  expect(result.msg.file.contentType).toEqual("text/plain");
  expect(result.msg.file.size).toEqual(5);
  expect(result.msg.file.filename).toEqual("my-file.txt");

  const contents = await result.msg.file.read();
  expect(contents.toString("utf-8")).toEqual("hello");

  const myfiles = await models.myFile.findMany({
    orderBy: { createdAt: "desc" },
  });

  const keys = myfiles.map((a) => a.file!.key);
  keys.push((result.msg.file as File).key);

  expect(myfiles.length).toEqual(3);
});

test("files - store once, write many", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.storeAndWriteMany({
    file: InlineFile.fromDataURL(dataUrl),
  });

  expect(result.msg.file.contentType).toEqual("text/plain");
  expect(result.msg.file.size).toEqual(5);
  expect(result.msg.file.filename).toEqual("my-file.txt");

  const contents = await result.msg.file.read();
  expect(contents.toString("utf-8")).toEqual("hello");

  const myfiles = await models.myFile.findMany();

  expect(myfiles.length).toEqual(3);

  // all files should have the same file key
  expect(myfiles[1].file?.key).toEqual(myfiles[0].file?.key);
  expect(myfiles[2].file?.key).toEqual(myfiles[0].file?.key);
});

test("files - model API file tests", async () => {
  await expect(actions.modelApiTests({})).not.toHaveError({});
});

test("files - kysely file tests", async () => {
  await expect(actions.kyselyTests({})).not.toHaveError({});
});

test("files - presigned url", async () => {
  const fileContents = "hello";
  const dataUrl = `data:text/plain;name=my-file.txt;base64,${Buffer.from(
    fileContents
  ).toString("base64")}`;

  const result = await actions.presignedUrl({
    file: InlineFile.fromDataURL(dataUrl),
  });
  const url = new URL(result);

  expect(url.searchParams.get("X-Amz-Algorithm")).toEqual("AWS4-HMAC-SHA256");
});

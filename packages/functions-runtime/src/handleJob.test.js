import { createJSONRPCRequest, JSONRPCErrorCode } from "json-rpc-2.0";
import { sql } from "kysely";
import { handleJob, RuntimeErrors } from "./handleJob";
import { test, expect, beforeEach, describe } from "vitest";
import { ModelAPI } from "./ModelAPI";
import { useDatabase } from "./database";
import KSUID from "ksuid";

test("when the job returns nothing as expected", async () => {
  const config = {
    jobs: {
      myJob: async (ctx, inputs) => {},
    },
    createJobContextAPI: () => {},
    permissions: {},
  };

  const rpcReq = createJSONRPCRequest("123", "myJob", { title: "a post" });
  rpcReq.meta = { permissionState: { status: "granted", reason: "role" } };

  expect(await handleJob(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    result: null,
  });
});

test("when there is an unexpected error in the job", async () => {
  const config = {
    jobs: {
      myJob: async (ctx, inputs) => {
        throw new Error("oopsie daisy");
      },
    },
    createJobContextAPI: () => {},
  };

  const rpcReq = createJSONRPCRequest("123", "myJob", { title: "a post" });
  rpcReq.meta = { permissionState: { status: "granted", reason: "role" } };

  expect(await handleJob(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.UnknownError,
      message: "oopsie daisy",
    },
  });
});

test("when there is an unexpected object thrown in the job", async () => {
  const config = {
    jobs: {
      myJob: async (ctx, inputs) => {
        throw { err: "oopsie daisy" };
      },
    },
    createJobContextAPI: () => {},
  };

  const rpcReq = createJSONRPCRequest("123", "myJob", { title: "a post" });
  rpcReq.meta = { permissionState: { status: "granted", reason: "role" } };

  expect(await handleJob(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.UnknownError,
      message: '{"err":"oopsie daisy"}',
    },
  });
});

test("when there is no matching job for the path", async () => {
  const config = {
    jobs: {
      myJob: async (ctx, inputs) => {},
    },
    createJobContextAPI: () => {},
  };

  const rpcReq = createJSONRPCRequest("123", "unknown", { title: "a post" });

  expect(await handleJob(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: JSONRPCErrorCode.MethodNotFound,
      message: "job 'unknown' does not exist or has not been implemented",
    },
  });
});

// The following tests assert on the various
// jsonrpc responses that *should* happen when a user
// writes a job that inadvertently causes a pg constraint error to occur inside of our ModelAPI class instance.
describe("ModelAPI error handling", () => {
  let functionConfig;
  let db;

  beforeEach(async () => {
    db = useDatabase();

    await sql`
      DROP TABLE IF EXISTS post;
      DROP TABLE IF EXISTS author;
  
      CREATE TABLE author(
        "id"               text PRIMARY KEY,
        "name"             text NOT NULL
      );
    
      CREATE TABLE post(
        "id"            text PRIMARY KEY,
        "title"         text NOT NULL UNIQUE,
        "author_id"     text NOT NULL REFERENCES author(id)
      );
      `.execute(db);

    await sql`
        INSERT INTO author (id, name) VALUES ('123', 'Bob')
      `.execute(db);

    const models = {
      post: new ModelAPI("post", undefined, {
        post: {
          author: {
            relationshipType: "belongsTo",
            foreignKey: "author_id",
            referencesTable: "person",
          },
        },
      }),
    };

    functionConfig = {
      jobs: {
        createPost: async (ctx, inputs) => {
          const post = await models.post.create({
            id: KSUID.randomSync().string,
            ...inputs,
          });
          return post;
        },
        deletePost: async (ctx, inputs) => {
          const deleted = await models.post.delete(inputs);
          return deleted;
        },
      },
      createJobContextAPI: () => ({}),
    };
  });

  test("when kysely returns a no result error", async () => {
    // a kysely NoResultError is thrown when attempting to delete/update a non existent record.
    const rpcReq = createJSONRPCRequest("123", "deletePost", {
      id: "non-existent-id",
    });
    rpcReq.meta = { permissionState: { status: "granted", reason: "role" } };

    expect(await handleJob(rpcReq, functionConfig)).toEqual({
      id: "123",
      jsonrpc: "2.0",
      error: {
        code: RuntimeErrors.RecordNotFoundError,
        message: "",
      },
    });
  });

  test("when there is a not null constraint error", async () => {
    const rpcReq = createJSONRPCRequest("123", "createPost", { title: null });
    rpcReq.meta = { permissionState: { status: "granted", reason: "role" } };

    expect(await handleJob(rpcReq, functionConfig)).toEqual({
      id: "123",
      jsonrpc: "2.0",
      error: {
        code: RuntimeErrors.NotNullConstraintError,
        message:
          'null value in column "title" of relation "post" violates not-null constraint',
        data: {
          code: "23502",
          column: "title",
          detail: expect.stringContaining("Failing row contains"),
          table: "post",
          value: undefined,
        },
      },
    });
  });

  test("when there is a uniqueness constraint error", async () => {
    await sql`
      INSERT INTO post (id, title, author_id) values(${
        KSUID.randomSync().string
      }, 'hello', '123')
      `.execute(db);

    const rpcReq = createJSONRPCRequest("123", "createPost", {
      title: "hello",
      author_id: "something",
    });
    rpcReq.meta = { permissionState: { status: "granted", reason: "role" } };

    expect(await handleJob(rpcReq, functionConfig)).toEqual({
      id: "123",
      jsonrpc: "2.0",
      error: {
        code: RuntimeErrors.UniqueConstraintError,
        message:
          'duplicate key value violates unique constraint "post_title_key"',
        data: {
          code: "23505",
          column: "title",
          detail: "Key (title)=(hello) already exists.",
          table: "post",
          value: "hello",
        },
      },
    });
  });

  test("when there is a null value in a foreign key column", async () => {
    const rpcReq = createJSONRPCRequest("123", "createPost", { title: "123" });
    rpcReq.meta = { permissionState: { status: "granted", reason: "role" } };

    expect(await handleJob(rpcReq, functionConfig)).toEqual({
      id: "123",
      jsonrpc: "2.0",
      error: {
        code: RuntimeErrors.NotNullConstraintError,
        message:
          'null value in column "author_id" of relation "post" violates not-null constraint',
        data: {
          code: "23502",
          column: "author_id",
          detail: expect.stringContaining("Failing row contains"),
          table: "post",
          value: undefined,
        },
      },
    });
  });

  test("when there is a foreign key constraint violation", async () => {
    const rpcReq = createJSONRPCRequest("123", "createPost", {
      title: "123",
      author_id: "fake",
    });
    rpcReq.meta = { permissionState: { status: "granted", reason: "role" } };

    expect(await handleJob(rpcReq, functionConfig)).toEqual({
      id: "123",
      jsonrpc: "2.0",
      error: {
        code: RuntimeErrors.ForeignKeyConstraintError,
        message:
          'insert or update on table "post" violates foreign key constraint "post_author_id_fkey"',
        data: {
          code: "23503",
          column: "author_id",
          detail: 'Key (author_id)=(fake) is not present in table "author".',
          table: "post",
          value: "fake",
        },
      },
    });
  });
});

import { createJSONRPCRequest, JSONRPCErrorCode } from "json-rpc-2.0";
import { sql } from "kysely";
import { handleRequest, RuntimeErrors } from "./handleRequest";
import { test, expect, beforeEach, describe } from "vitest";
import { ModelAPI } from "./ModelAPI";
import { useDatabase } from "./database";
const { Permissions } = require("./permissions");
import { PROTO_ACTION_TYPES } from "./consts";
import KSUID from "ksuid";

test("when the custom function returns expected value", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs) => {
        new Permissions().allow();

        return {
          title: "a post",
          id: "abcde",
        };
      },
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {
      return {
        response: {
          headers: new Headers(),
        },
      };
    },
  };

  const rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    meta: {
      headers: {},
    },
    result: {
      title: "a post",
      id: "abcde",
    },
  });
});

test("when there is no matching function for the path", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs) => {},
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {
      return {
        response: {
          headers: new Headers(),
        },
      };
    },
  };

  const rpcReq = createJSONRPCRequest("123", "unknown", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: JSONRPCErrorCode.MethodNotFound,
      message: "no corresponding function found for 'unknown'",
    },
  });
});

test("when there is an unexpected error in the custom function", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs) => {
        throw new Error("oopsie daisy");
      },
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {
      return {
        response: {
          headers: new Headers(),
        },
      };
    },
  };

  const rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.UnknownError,
      message: "oopsie daisy",
    },
  });
});

test("when a role based permission has already been granted by the main runtime", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs, api) => {
        return {
          title: inputs.title,
        };
      },
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createModelAPI: () => {},
    createContextAPI: () => {
      return {
        response: {
          headers: new Headers(),
        },
      };
    },
  };

  let rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  Object.assign(rpcReq, {
    ...rpcReq,
    meta: { permissionState: { status: "granted", reason: "role" } },
  });
  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    result: {
      title: "a post",
    },
    meta: {
      headers: {},
    },
  });
});

test("when there is an unexpected object thrown in the custom function", async () => {
  const config = {
    functions: {
      createPost: async (ctx, inputs) => {
        throw { err: "oopsie daisy" };
      },
    },
    actionTypes: {
      createPost: PROTO_ACTION_TYPES.CREATE,
    },
    createContextAPI: () => {
      return {
        response: {
          headers: new Headers(),
        },
      };
    },
  };

  const rpcReq = createJSONRPCRequest("123", "createPost", { title: "a post" });

  expect(await handleRequest(rpcReq, config)).toEqual({
    id: "123",
    jsonrpc: "2.0",
    error: {
      code: RuntimeErrors.UnknownError,
      message: '{"err":"oopsie daisy"}',
    },
  });
});

// The following tests assert on the various
// jsonrpc responses that *should* happen when a user
// writes a custom function that inadvertently causes a pg constraint error to occur inside of our ModelAPI class instance.
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
      INSERT INTO author (id, name) VALUES ('adam', 'adam bull')
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
      permissionFns: {},
      actionTypes: {
        createPost: PROTO_ACTION_TYPES.CREATE,
        deletePost: PROTO_ACTION_TYPES.DELETE,
      },
      functions: {
        createPost: async (ctx, inputs) => {
          new Permissions().allow();

          const post = await models.post.create({
            id: KSUID.randomSync().string,
            ...inputs,
          });

          return post;
        },
        deletePost: async (ctx, inputs) => {
          new Permissions().allow();

          const deleted = await models.post.delete(inputs);

          return deleted;
        },
      },
      createContextAPI: () => {
        return {
          response: {
            headers: new Headers(),
          },
        };
      },
    };
  });

  test("when kysely returns a no result error", async () => {
    // a kysely NoResultError is thrown when attempting to delete/update a non existent record.
    const rpcReq = createJSONRPCRequest("123", "deletePost", {
      id: "non-existent-id",
    });

    expect(await handleRequest(rpcReq, functionConfig)).toEqual({
      id: "123",
      jsonrpc: "2.0",
      error: {
        code: RuntimeErrors.RecordNotFoundError,
        message: "no result",
      },
    });
  });

  test("when there is a not null constraint error", async () => {
    const rpcReq = createJSONRPCRequest("123", "createPost", { title: null });

    expect(await handleRequest(rpcReq, functionConfig)).toEqual({
      id: "123",
      jsonrpc: "2.0",
      error: {
        code: RuntimeErrors.NotNullConstraintError,
        message: 'null value in column "title" violates not-null constraint',
        data: {
          code: "23502",
          column: "title",
          detail: expect.stringContaining("Failing row contains"),
          table: "post",
        },
      },
    });
  });
  test("when there is a uniqueness constraint error", async () => {
    await sql`
    INSERT INTO post (id, title, author_id) values(${
      KSUID.randomSync().string
    }, 'hello', 'adam')
    `.execute(db);

    const rpcReq = createJSONRPCRequest("123", "createPost", {
      title: "hello",
      author_id: "something",
    });

    expect(await handleRequest(rpcReq, functionConfig)).toEqual({
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

    expect(await handleRequest(rpcReq, functionConfig)).toEqual({
      id: "123",
      jsonrpc: "2.0",
      error: {
        code: RuntimeErrors.NotNullConstraintError,
        message:
          'null value in column "author_id" violates not-null constraint',
        data: {
          code: "23502",
          column: "author_id",
          detail: expect.stringContaining("Failing row contains"),
          table: "post",
        },
      },
    });
  });

  test("when there is a foreign key constraint violation", async () => {
    const rpcReq2 = createJSONRPCRequest("123", "createPost", {
      title: "123",
      author_id: "fake",
    });

    expect(await handleRequest(rpcReq2, functionConfig)).toEqual({
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

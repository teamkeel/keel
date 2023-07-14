package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/proto"
)

func GenerateClient(ctx context.Context, schema *proto.Schema, makePackage bool, apiName string) (codegen.GeneratedFiles, error) {

	api := schema.Apis[0]

	if apiName != "" {
		match := false
		for _, a := range schema.Apis {
			if strings.EqualFold(a.Name, apiName) {
				match = true
				api = a
			}
		}
		if !match {
			return nil, fmt.Errorf("No %s API found", apiName)
		}
	}

	var files codegen.GeneratedFiles

	if makePackage {
		files = generateClientSdkPackage(schema, api)
		return files, nil
	}

	files = generateClientSdkFile(schema, api)
	return files, nil
}

// Break this up so that we can generate either a single file client or a full package (package can use cross-fetch)

func generateClientSdkFile(schema *proto.Schema, api *proto.Api) codegen.GeneratedFiles {
	client := &codegen.Writer{}

	client.Writeln("// GENERATED DO NOT EDIT")
	client.Writeln("")

	client.Writeln(clientCore)
	client.Writeln(clientTypes)

	client.Writeln("")
	client.Writeln("// API")
	client.Writeln("")

	writeClientAPIClass(client, schema, api)

	return []*codegen.GeneratedFile{
		{
			Path:     "keelClient.ts",
			Contents: client.String(),
		},
	}
}

func generateClientSdkPackage(schema *proto.Schema, api *proto.Api) codegen.GeneratedFiles {
	core := &codegen.Writer{}
	client := &codegen.Writer{}
	types := &codegen.Writer{}

	core.Writeln(`import fetch from "cross-fetch"`)
	core.Writeln(`import { APIError, APIResult } from "./types";`)
	core.Writeln("")
	core.Writeln(clientCore)

	types.Writeln(clientTypes)

	client.Writeln(`import { CoreClient, RequestConfig } from "./core";`)
	client.Writeln("")
	writeClientAPIClass(client, schema, api)

	return []*codegen.GeneratedFile{
		{
			Path:     "@teamkeel/client/core.ts",
			Contents: core.String(),
		},
		{
			Path:     "@teamkeel/client/index.ts",
			Contents: client.String(),
		},
		{
			Path:     "@teamkeel/client/types.ts",
			Contents: types.String(),
		},
		{
			Path: "@teamkeel/client/package.json",
			Contents: `{
	"name": "@teamkeel/client",
	"dependencies": {
		"cross-fetch": "4.0.0"
	}
}`,
		},
	}
}

func writeClientAPIClass(w *codegen.Writer, schema *proto.Schema, api *proto.Api) {

	w.Writeln("export class APIClient extends CoreClient {")

	w.Indent()
	w.Writeln(`constructor(config: RequestConfig) {
		super(config);
	}`)

	apiModels := lo.Map(api.ApiModels, func(a *proto.ApiModel, index int) string {
		return a.ModelName
	})

	for _, model := range schema.Models {

		// Skip any models not part of this api
		if !lo.Contains(apiModels, model.Name) {
			continue
		}

		for _, op := range model.Operations {
			msg := proto.FindMessage(schema.Messages, op.InputMessageName)

			w.Writef("%s(i", op.Name)

			// Check that all of the top level fields in the matching message are optional
			// If so, then we can make it so you don't even need to specify the key
			// example, this allows for:
			// await actions.listActivePublishersWithActivePosts();
			// instead of:
			// const { results: publishers } =
			// await actions.listActivePublishersWithActivePosts({ where: {} });
			if lo.EveryBy(msg.Fields, func(f *proto.MessageField) bool {
				return f.Optional
			}) {
				w.Write("?")
			}

			w.Writef(`: %s) `, op.InputMessageName)
			w.Writeln("{")

			w.Indent()
			w.Writef(`return this.request<%s>("%s", i)`, toClientActionReturnType(model, op), op.Name)
			w.Writeln(";")
			w.Dedent()
			w.Writeln("}")
		}
	}
	w.Dedent()
	w.Writeln("}")

	w.Writeln("")
	w.Writeln("// API Types")
	w.Writeln("")

	writeMessages(w, schema, false, true)

	for _, enum := range schema.Enums {
		writeEnum(w, enum)
		writeEnumWhereCondition(w, enum)
		writeEnumObject(w, enum)
	}

	for _, model := range schema.Models {

		// Skip any models not part of this api
		if !lo.Contains(apiModels, model.Name) {
			continue
		}

		writeModelInterface(w, model, true)
	}

}

func toClientActionReturnType(model *proto.Model, op *proto.Operation) string {
	returnType := ""
	sdkPrefix := ""

	switch op.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		returnType += sdkPrefix + model.Name
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		returnType += sdkPrefix + model.Name
	case proto.OperationType_OPERATION_TYPE_GET:
		returnType += sdkPrefix + model.Name + " | null"
	case proto.OperationType_OPERATION_TYPE_LIST:
		returnType += "{results: " + sdkPrefix + model.Name + "[], pageInfo: any}"
	case proto.OperationType_OPERATION_TYPE_DELETE:
		// todo: create ID type
		returnType += "string"
	case proto.OperationType_OPERATION_TYPE_READ, proto.OperationType_OPERATION_TYPE_WRITE:
		returnType += op.ResponseMessageName
	}

	returnType += ""
	return returnType
}

var clientCore = `type RequestHeaders = Record<string, string>;

export type RequestConfig = {
  endpoint: string;
  headers?: RequestHeaders;
};

export class CoreClient {
  private token = "";
  constructor(private config: RequestConfig) {}

  _setHeaders(headers: RequestHeaders): CoreClient {
    this.config.headers = headers;
    return this;
  }

  _setHeader(key: string, value: string): CoreClient {
    const { headers } = this.config;
    if (headers) {
      headers[key] = value;
    } else {
      this.config.headers = { [key]: value };
    }
    return this;
  }

  _setEndpoint(value: string): CoreClient {
    this.config.endpoint = value;
    return this;
  }

  _setToken(value: string): CoreClient {
    this.token = value;
    return this;
  }

  _clearToken(): CoreClient {
    this.token = "";
    return this;
  }

  async request<T>(action: string, body: any): Promise<APIResult<T>> {
    const res = fetch(
      stripTrailingSlash(this.config.endpoint) + "/json/" + action,
      {
        method: "POST",
        cache: "no-cache",
        headers: {
          accept: "application/json",
          "content-type": "application/json",
          ...this.config.headers,
          ...(this.token
            ? {
                Authorization: "Bearer " + this.token,
              }
            : {}),
        },
        body: JSON.stringify(body),
      }
    );

    res.catch((err) => {
      return {
        error: {
          type: "unknown",
          message: "unknown error",
          err,
        },
      };
    });

    const result = await res;

    if (result.status >= 200 && result.status < 299) {
      return {
        data: await result.json(),
      };
    }

    let errorMessage = "unknown error";

    try {
      const errorData: {
        message: string;
      } = await result.json();
      errorMessage = errorData.message;
    } catch (error) {}

    const requestId = result.headers.get("X-Amzn-Requestid") || undefined;

    const errorCommon = {
      message: errorMessage,
      requestId,
    };

    switch (result.status) {
      case 400:
        return {
          error: {
            ...errorCommon,
            type: "bad_request",
          },
        };
      case 401:
        return {
          error: {
            ...errorCommon,
            type: "unauthorized",
          },
        };
      case 403:
        return {
          error: {
            ...errorCommon,
            type: "forbidden",
          },
        };
      case 404:
        return {
          error: {
            ...errorCommon,
            type: "not_found",
          },
        };
      case 500:
        return {
          error: {
            ...errorCommon,
            type: "internal_server_error",
          },
        };

      default:
        return {
          error: {
            ...errorCommon,
            type: "unknown",
          },
        };
    }
  }
}

// Utils

const stripTrailingSlash = (str: string) => {
  return str.endsWith("/") ? str.slice(0, -1) : str;
};
`

var clientTypes = `// Result type

export type APIResult<T> = Promise<Result<T, APIError>>;

type Data<T> = {
  data: T;
  error?: never;
};

type Err<U> = {
  data?: never;
  error: U;
};

type Result<T, U> = NonNullable<Data<T> | Err<U>>;

// Error types

/* 400 */
type BadRequestError = {
  type: "bad_request";
  message: string;
  requestId?: string;
};

/* 401 */
type UnauthorizedError = {
  type: "unauthorized";
  message: string;
  requestId?: string;
};

/* 403 */
type ForbiddenError = {
  type: "forbidden";
  message: string;
  requestId?: string;
};

/* 404 */
type NotFoundError = {
  type: "not_found";
  message: string;
  requestId?: string;
};

/* 500 */
type InternalServerError = {
  type: "internal_server_error";
  message: string;
  requestId?: string;
};

/* Unhandled/unexpected errors */
type UnknownError = {
  type: "unknown";
  message: string;
  err?: unknown;
  requestId?: string;
};

export type APIError =
  | UnauthorizedError
  | ForbiddenError
  | NotFoundError
  | BadRequestError
  | InternalServerError
  | UnknownError;
`

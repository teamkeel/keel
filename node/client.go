package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema/parser"
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
			return nil, fmt.Errorf("api not found: %s", apiName)
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
	client.Writeln(authTypes)

	client.Writeln("")
	client.Writeln("// API")
	client.Writeln("")

	writeClientApiClass(client, schema, api)

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

	client.Writeln(`import { Core, RequestConfig } from "./core";`)
	client.Writeln("")
	writeClientApiClass(client, schema, api)

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

func writeClientApiClass(w *codegen.Writer, schema *proto.Schema, api *proto.Api) {
	w.Writeln("export class APIClient extends Core {")

	w.Indent()
	w.Writeln(`constructor(config: RequestConfig) {
		super(config);
    }`)

	w.Writeln("private actions = {")
	w.Indent()

	writeClientActions(w, schema, api)

	w.Dedent()
	w.Writeln("};")
	w.Writeln("")

	w.Writeln("api = {")
	w.Indent()

	writeClientApiDefinition(w, schema, api)

	w.Dedent()
	w.Writeln("};")

	w.Dedent()
	w.Writeln("}")

	w.Writeln("")

	writeClientTypes(w, schema, api)
}

func writeClientActions(w *codegen.Writer, schema *proto.Schema, api *proto.Api) {
	for _, a := range proto.GetActionNamesForApi(schema, api) {
		action := proto.FindAction(schema, a)
		msg := proto.FindMessage(schema.Messages, action.InputMessageName)

		w.Writef("%s: (i", action.Name)

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

		inputType := action.InputMessageName
		if inputType == parser.MessageFieldTypeAny {
			inputType = "any"
		}

		w.Writef(`: %s) `, inputType)
		w.Writeln("=> {")

		w.Indent()

		model := proto.FindModel(schema.Models, action.ModelName)
		w.Writef(`return this.client.rawRequest<%s>("%s", i)`, toClientActionReturnType(model, action), action.Name)

		var setTokenChain = `.then((res) => {
				if (res.data && res.data.token) this.client.setToken(res.data.token);
				return res;
            })`

		if action.Name == "authenticate" {
			w.Writef(setTokenChain)
		}

		w.Writeln(";")
		w.Dedent()
		w.Writeln("},")
	}

}

func writeClientApiDefinition(w *codegen.Writer, schema *proto.Schema, api *proto.Api) {
	queries := []string{}
	mutations := []string{}

	for _, a := range proto.GetActionNamesForApi(schema, api) {
		action := proto.FindAction(schema, a)
		if action.Type == proto.ActionType_ACTION_TYPE_GET || action.Type == proto.ActionType_ACTION_TYPE_LIST || action.Type == proto.ActionType_ACTION_TYPE_READ {
			queries = append(queries, action.Name)
		} else {
			mutations = append(mutations, action.Name)
		}

	}

	w.Writeln("queries: {")
	w.Indent()
	for _, fn := range queries {
		w.Writef(`%s: this.actions.%s`, fn, fn)
		w.Writeln(",")
	}
	w.Dedent()
	w.Writeln("},")

	w.Writeln("mutations: {")
	w.Indent()
	for _, fn := range mutations {
		w.Writef(`%s: this.actions.%s`, fn, fn)
		w.Writeln(",")
	}
	w.Dedent()
	w.Writeln("}")
}

func writeClientTypes(w *codegen.Writer, schema *proto.Schema, api *proto.Api) {
	w.Writeln("// API Types")
	w.Writeln("")

	writeMessages(w, schema, false)

	for _, enum := range schema.Enums {
		writeEnum(w, enum)
		writeEnumWhereCondition(w, enum)
	}

	models := proto.ApiModels(schema, api)

	for _, model := range models {
		writeModelInterface(w, model)
	}

	w.Writeln(`export type SortDirection = "asc" | "desc" | "ASC" | "DESC";`)

	w.Writeln("")
	w.Writeln("type PageInfo = {")
	w.Indent()
	w.Writeln("count: number;")
	w.Writeln("endCursor: string;")
	w.Writeln("hasNextPage: boolean;")
	w.Writeln("startCursor: string;")
	w.Writeln("totalCount: number;")
	w.Dedent()
	w.Writeln("};")

}

func toClientActionReturnType(model *proto.Model, op *proto.Action) string {
	switch op.Type {
	case proto.ActionType_ACTION_TYPE_CREATE:
		return model.Name
	case proto.ActionType_ACTION_TYPE_UPDATE:
		return model.Name
	case proto.ActionType_ACTION_TYPE_GET:
		return model.Name + " | null"
	case proto.ActionType_ACTION_TYPE_LIST:
		return "{results: " + model.Name + "[], pageInfo: PageInfo}"
	case proto.ActionType_ACTION_TYPE_DELETE:
		return "string"
	case proto.ActionType_ACTION_TYPE_READ, proto.ActionType_ACTION_TYPE_WRITE:
		if op.ResponseMessageName == parser.MessageFieldTypeAny {
			return "any"
		}

		return op.ResponseMessageName
	default:
		panic(fmt.Sprintf("unexpected action type: %s", op.Type.String()))
	}
}

var clientCore = `type RequestHeaders = globalThis.Record<string, string>;

export type RequestConfig = {
  baseUrl: string;
  headers?: RequestHeaders;
};

export class Core {
	constructor(private config: RequestConfig) {}

	session: Session | null = null;

	ctx = {
		/**
		 * @deprecated This has been deprecated in favour of APIClient.auth.getSession()
		 */
		token: "",
		isAuthenticated: false,
	};

	client = {
    setHeaders: (headers: RequestHeaders): Core => {
      this.config.headers = headers;
      return this;
    },
    setHeader: (key: string, value: string): Core => {
      const { headers } = this.config;
      if (headers) {
        headers[key] = value;
      } else {
        this.config.headers = { [key]: value };
      }
      return this;
    },
    setBaseUrl: (value: string): Core => {
      this.config.baseUrl = value;
      return this;
    },
    /**
     * @deprecated This has been deprecated in favour of the APIClient.auth authenticate 
     * helper functions or APIClient.auth.setSession()
     */
    setToken: (value: string): Core => {
      this.ctx.token = value;
      this.ctx.isAuthenticated = true;
      return this;
    },
    /**
     * @deprecated This has been deprecated in favour of APIClient.auth.logout()
     */
    clearToken: (): Core => {
      this.ctx.token = "";
      this.ctx.isAuthenticated = false;
      return this;
    },
    rawRequest: async <T>(action: string, body: any): Promise<APIResult<T>> => {
      try {
        const result = await globalThis.fetch(
          stripTrailingSlash(this.config.baseUrl) + "/json/" + action,
          {
            method: "POST",
            cache: "no-cache",
            headers: {
              accept: "application/json",
              "content-type": "application/json",
              ...this.config.headers,
              ...(this.ctx.token
                ? {
                    Authorization: "Bearer " + this.ctx.token,
                  }
                : {}),
            },
            body: JSON.stringify(body),
          }
        );

        if (result.status >= 200 && result.status < 299) {
          const rawJson = await result.text();
          const data = JSON.parse(rawJson, reviver);

          return {
            data,
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
      } catch (error) {
        return {
          error: {
            type: "unknown",
            message: "unknown error",
            error,
          },
        };
      }
    },
  };

  auth = {
    // Manually set the session with an access and refresh token acquired elsewhere.
    // We recommend to rather use the built-in authentication functions provided in this client.
    setSession: (accessToken: string, refreshToken: string) => {
      this.session = { accessToken, refreshToken };
    },

    // Retrieves the current session's access and refresh tokens. This will also refresh the 
    // session with the authentication server.
    getSession: async () => {
      await this.auth.refreshSession();
      return this.session;
    },

    // Returns the list of supported auth providers and their SSO login URLs.
    providers: async (): Promise<Provider[]> => {
      const authUrl = this.config.baseUrl.replace("/api", "");

      const result = await globalThis.fetch(
        stripTrailingSlash(authUrl) + "/auth/providers",
        {
          method: "GET",
          cache: "no-cache",
          headers: {
            "content-type": "application/json",
          },
        },
      );

      if (result.status == 200) {
        const rawJson = await result.text();
        return JSON.parse(rawJson);
      } else {
        throw new Error("unexpected status code response from /auth/providers: " + result.status)
      }
    },

    // Authenticate with an OIDC token.
    authenticateWithIdToken: async (idToken: string) => {
      const req: TokenExchangeGrant = {
        grant: "token_exchange",
        subjectToken: idToken
      }

      await this.auth.tokenRequest(req)
    },

    // Authenticate with an SSO Login code.
    authenticateWithSSOLoginCode: async (code: string) => {
      const req: AuthorizationCodeGrant = {
        grant: "authorization_code",
        code: code
      }

      await this.auth.tokenRequest(req)
    },

	// Refresh the session with the authentication server.
    refreshSession: async () => {
      if (!this.session) {
        return;
      }

      const req: RefreshGrant = {
        grant:  "refresh_token",
        refreshToken: this.session!.refreshToken
      }

      await this.auth.tokenRequest(req);
    },

    tokenRequest: async(req: TokenGrant) => {
      let body = null;      
      switch (req.grant) {
        case "token_exchange":
          body = {
            "grant_type": req.grant,
            "subject_token": req.subjectToken
          };
          break;
        case "authorization_code":
          body = {
            "grant_type": req.grant,
            "code": req.code
          };
          break;
        case "refresh_token":
          body = {
            "grant_type": req.grant,
            "refresh_token": req.refreshToken
          };
          break;
        default:
          throw new Error("grant not implemented")
      }

      const authUrl = this.config.baseUrl.replace("/api", "");
      const result = await globalThis.fetch(
        stripTrailingSlash(authUrl) + "/auth/token",
        {
          method: "POST",
          cache: "no-cache",
          headers: {
            accept: "application/json",
            "content-type": "application/json",
          },
          body: JSON.stringify(body),
        },
      );
     
      if (result.status == 200) {
        const rawJson = await result.text();
        const data = JSON.parse(rawJson);
        this.auth.setSession(data.access_token, data.refresh_token);
        this.client.setToken(data.access_token); // Deprecate
      } else {
        this.auth.logout();
        this.client.clearToken(); // Deprecate

        const resp = await result.json();
        throw new TokenError(resp.error, resp.error_description);
      }
    },

    logout: async () => {
        const authUrl = this.config.baseUrl.replace("/api", "");
        const result = await globalThis.fetch(
          stripTrailingSlash(authUrl) + "/auth/revoke",
          {
            method: "POST",
            cache: "no-cache",
            headers: {
              accept: "application/json",
              "content-type": "application/json",
            },
            body: JSON.stringify({
              token: this.session?.refreshToken,
            }),
          },
        );
  
        this.session = null;
        this.client.clearToken();
    },
  };
}

// Utils

const stripTrailingSlash = (str: string) => {
  if (!str) return str;
  return str.endsWith("/") ? str.slice(0, -1) : str;
};

const RFC3339 = /^(?:\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12][0-9]|3[01]))?(?:[T\s](?:[01]\d|2[0-3]):[0-5]\d(?::[0-5]\d)?(?:\.\d+)?(?:[Zz]|[+-](?:[01]\d|2[0-3]):?[0-5]\d)?)?$/;
function reviver(key: any, value: any) {
  // Convert any ISO8601/RFC3339 strings to dates
  if (value && typeof value === "string" && RFC3339.test(value)) {
	return new Date(value);
  }
  return value;
}

`

var clientTypes = `// Result types

export type APIResult<T> = Result<T, APIError>;

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
  error?: unknown;
  requestId?: string;
};

export type APIError =
  | UnauthorizedError
  | ForbiddenError
  | NotFoundError
  | BadRequestError
  | InternalServerError
  | UnknownError;

// Auth

export interface Provider {
  name: string;
  type: string;
  authorizeUrl: string;
};

export interface Session {
  accessToken: string;
  refreshToken: string;
};

export type TokenExchangeGrant = {
  grant: "token_exchange";
  subjectToken: string;
};

export type AuthorizationCodeGrant = {
  grant: "authorization_code";
  code: string;
};

export type RefreshGrant = {
  grant: "refresh_token";
  refreshToken: string;
};

export type TokenGrant = 
  | TokenExchangeGrant 
  | AuthorizationCodeGrant 
  | RefreshGrant;

export class TokenError extends Error {
  errorDescription: string;
  constructor( error: string, errorDescription: string ) {
      super();
      this.message = error;
      this.errorDescription = errorDescription;
  }
}
`

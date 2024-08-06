package node

import (
	"context"
	"fmt"
	"io/fs"
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

	writeClientStandardTypes(client)
	writeClientCore(client)

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

func writeClientStandardTypes(w *codegen.Writer) {
	b, _ := fs.ReadFile(templates, "templates/client/types.d.ts")
	w.Writeln(string(b))
}

func writeClientCore(w *codegen.Writer) {
	b, _ := fs.ReadFile(templates, "templates/client/core.ts")
	w.Writeln(string(b))
}

func generateClientSdkPackage(schema *proto.Schema, api *proto.Api) codegen.GeneratedFiles {
	core := &codegen.Writer{}
	client := &codegen.Writer{}
	types := &codegen.Writer{}

	core.Writeln(`import fetch from "cross-fetch"`)
	core.Writeln(`import { APIError, APIResult } from "./types";`)
	core.Writeln("")
	writeClientStandardTypes(core)

	writeClientCore(types)

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
		"cross-fetch": "^4.0.0"
	}
}`,
		},
	}
}

func writeClientApiClass(w *codegen.Writer, schema *proto.Schema, api *proto.Api) {
	w.Writeln("export class APIClient extends Core {")

	w.Indent()
	w.Writeln("constructor(config: Config) {")
	w.Indent()
	w.Writeln("super(config);")
	w.Dedent()
	w.Writeln("}")
	w.Writeln("")

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

	// writing embedded response types
	for _, a := range proto.GetActionNamesForApi(schema, api) {
		action := proto.FindAction(schema, a)
		embeds := action.GetResponseEmbeds()
		if len(embeds) == 0 {
			continue
		}
		model := proto.FindModel(schema.Models, action.ModelName)
		writeEmbeddedModelInterface(w, schema, model, toResponseType(action.Name), embeds)
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

	w.Writeln(`export declare class InlineFile {`)
	w.Indent()
	w.Writeln(`constructor(key: any, filename: any, contentType: any, size: any, url: any);`)
	w.Writeln(`static fromObject(obj: any): InlineFile;`)
	w.Writeln(`static fromDataURL(url: string): InlineFile;`)
	w.Writeln(`read(): Buffer;`)
	w.Writeln(`store(expires?: Date): Promise<any>;`)
	w.Writeln(`filename: string;`)
	w.Writeln(`contentType: string;`)
	w.Writeln(`size: number;`)
	w.Writeln(`url: string | null;`)
	w.Dedent()
	w.Writeln(`}`)

}

func toClientActionReturnType(model *proto.Model, op *proto.Action) string {
	switch op.Type {
	case proto.ActionType_ACTION_TYPE_CREATE:
		return model.Name
	case proto.ActionType_ACTION_TYPE_UPDATE:
		return model.Name
	case proto.ActionType_ACTION_TYPE_GET:
		if len(op.GetResponseEmbeds()) > 0 {
			return toResponseType(op.Name) + " | null"
		}
		return model.Name + " | null"
	case proto.ActionType_ACTION_TYPE_LIST:
		respName := model.Name
		if len(op.GetResponseEmbeds()) > 0 {
			respName = toResponseType(op.Name)
		}
		return "{results: " + respName + "[], pageInfo: PageInfo}"
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

package node

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/iancoleman/strcase"
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
	writeClientApiInterface(w, schema, api)

	w.Writeln("export class APIClient extends Core {")

	w.Indent()
	w.Writeln("constructor(config: Config) {")
	w.Indent()
	w.Writeln("super(config);")
	w.Dedent()
	w.Writeln("}")
	w.Writeln("")

	w.Writeln("api = {")
	w.Indent()
	w.Writeln("queries: new Proxy({}, {")
	w.Indent()
	w.Writeln("get: (_, fn: string) => (i: any) => this.client.rawRequest(fn, i),")
	w.Dedent()
	w.Writeln("}),")
	w.Writeln("mutations: new Proxy({}, {")
	w.Indent()
	w.Writeln("get: (_, fn: string) => (i: any) => this.client.rawRequest(fn, i),")
	w.Dedent()
	w.Writeln("})")
	w.Dedent()
	w.Writeln("} as KeelAPI;")

	w.Dedent()
	w.Writeln("}")

	w.Writeln("")

	writeClientTypes(w, schema, api)
}

func writeClientApiInterface(w *codegen.Writer, schema *proto.Schema, api *proto.Api) {
	queries := []*proto.Action{}
	mutations := []*proto.Action{}

	for _, a := range proto.GetActionNamesForApi(schema, api) {
		action := schema.FindAction(a)
		if action.Type == proto.ActionType_ACTION_TYPE_GET || action.Type == proto.ActionType_ACTION_TYPE_LIST || action.Type == proto.ActionType_ACTION_TYPE_READ {
			queries = append(queries, action)
		} else {
			mutations = append(mutations, action)
		}
	}

	w.Writeln("interface KeelAPI {")
	w.Indent()
	w.Writeln("queries: {")
	w.Indent()

	for _, fn := range queries {
		writeClientActionType(w, schema, fn)
	}

	w.Dedent()
	w.Writeln("},")

	w.Writeln("mutations: {")
	w.Indent()

	for _, fn := range mutations {
		writeClientActionType(w, schema, fn)
	}

	w.Dedent()
	w.Writeln("}")

	w.Dedent()
	w.Writeln("}")
	w.Writeln("")
}

func writeClientActionType(w *codegen.Writer, schema *proto.Schema, action *proto.Action) {
	msg := schema.FindMessage(action.InputMessageName)

	w.Writef("%s: (i", action.Name)

	// Check that all of the top level fields in the matching message are optional
	if lo.EveryBy(msg.Fields, func(f *proto.MessageField) bool {
		return f.Optional
	}) {
		w.Write("?")
	}

	inputType := action.InputMessageName
	if inputType == parser.MessageFieldTypeAny {
		inputType = "any"
	}

	w.Writef(`: %s) => `, inputType)

	model := schema.FindModel(action.ModelName)
	w.Writef(`Promise<APIResult<%s>>`, toClientActionReturnType(model, action))

	w.Writeln(";")
}

func writeClientTypes(w *codegen.Writer, schema *proto.Schema, api *proto.Api) {
	w.Writeln("// API Types")
	w.Writeln("")

	writeMessages(w, schema, false, true)

	for _, enum := range schema.Enums {
		writeEnum(w, enum)
		writeEnumWhereCondition(w, enum)
	}

	models := proto.ApiModels(schema, api)

	for _, model := range models {
		writeModelInterface(w, model, true)
	}

	for _, a := range proto.GetActionNamesForApi(schema, api) {
		action := schema.FindAction(a)
		writeResultInfoInterface(w, schema, action, true)
	}

	// writing embedded response types
	for _, a := range proto.GetActionNamesForApi(schema, api) {
		action := schema.FindAction(a)
		embeds := action.GetResponseEmbeds()
		if len(embeds) == 0 {
			continue
		}
		model := schema.FindModel(action.ModelName)
		writeEmbeddedModelInterface(w, schema, model, toResponseType(action.Name), embeds)
	}

	w.Writeln("")
}

func toClientActionReturnType(model *proto.Model, action *proto.Action) string {
	switch action.Type {
	case proto.ActionType_ACTION_TYPE_CREATE:
		return model.Name
	case proto.ActionType_ACTION_TYPE_UPDATE:
		return model.Name
	case proto.ActionType_ACTION_TYPE_GET:
		if len(action.GetResponseEmbeds()) > 0 {
			return toResponseType(action.Name) + " | null"
		}
		return model.Name + " | null"
	case proto.ActionType_ACTION_TYPE_LIST:
		respName := model.Name
		if len(action.GetResponseEmbeds()) > 0 {
			respName = toResponseType(action.Name)
		}

		if len(action.Facets) > 0 {
			resultInfo := fmt.Sprintf("%sResultInfo", strcase.ToCamel(action.Name))
			return "{ results: " + respName + "[], resultInfo: " + resultInfo + ", pageInfo: PageInfo }"
		} else {
			return "{ results: " + respName + "[], pageInfo: PageInfo }"
		}
	case proto.ActionType_ACTION_TYPE_DELETE:
		return "string"
	case proto.ActionType_ACTION_TYPE_READ, proto.ActionType_ACTION_TYPE_WRITE:
		if action.ResponseMessageName == parser.MessageFieldTypeAny {
			return "any"
		}

		return action.ResponseMessageName
	default:
		panic(fmt.Sprintf("unexpected action type: %s", action.Type.String()))
	}
}

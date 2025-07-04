package node

import (
	"fmt"
	"regexp"

	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/permissions"
	"github.com/teamkeel/keel/proto"
)

var (
	sqlPlaceholderRegexp = regexp.MustCompile(`\?`)
)

// writePermissions writes a JS object where the keys are function names
// and the values a list of functions that can be run to check permissions
// for a list of records.
func writePermissions(w *codegen.Writer, schema *proto.Schema) {
	w.Writeln("export const permissionFns = {")
	w.Indent()

	for _, model := range schema.GetModels() {
		for _, action := range model.GetActions() {
			if action.GetImplementation() != proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM {
				continue
			}

			perms := proto.PermissionsForAction(schema, action)

			// TODO: think about how to handle error's here
			sql, values, _ := permissions.ToSQL(schema, model, perms)
			if sql == "" {
				w.Writef("%s: [],\n", action.GetName())
				continue
			}

			w.Writef("%s: [\n", action.GetName())
			w.Indent()

			w.Writeln("async (records, ctx, db) => {")
			w.Indent()

			w.Write("const { rows } = await sql`")
			valueIdx := 0

			// Kysely uses JS interpolation rather than placeholders in the query,
			// so we replace all the "?"'s with JS interpolations of the appropriate value
			sql = sqlPlaceholderRegexp.ReplaceAllStringFunc(sql, func(_ string) string {
				v := values[valueIdx]
				valueIdx++

				switch v.Type {
				case permissions.ValueIdentityID:
					return "${ctx.identity ? ctx.identity.id : ''}"
				case permissions.ValueIdentityEmail:
					return "${ctx.identity ? ctx.identity.email : ''}"
				case permissions.ValueNow:
					return "${ctx.now()}"
				case permissions.ValueIsAuthenticated:
					return "${ctx.isAuthenticated}"
				case permissions.ValueRecordIDs:
					// Need to use sql.join() here:
					// Docs: https://kysely-org.github.io/kysely/interfaces/Sql.html#join
					// Note: sql.join annoyingly throws an exception for zero length arrays
					return "${(records.length > 0) ? sql.join(records.map(x => x.id)) : []}"
				case permissions.ValueString:
					// Note: StringValue is already wrapped in double quotes
					return fmt.Sprintf(`${%s}`, v.StringValue)
				case permissions.ValueNumber:
					return fmt.Sprintf(`${%d}`, v.NumberValue)
				case permissions.ValueHeader:
					return fmt.Sprintf(`${ctx.headers["%s"] || ""}`, v.HeaderKey)
				case permissions.ValueSecret:
					return fmt.Sprintf(`${ctx.secrets["%s"] || ""}`, v.SecretKey)
				}

				return ""
			})
			w.Write(sql)
			w.Writeln("`.execute(db);")

			// Permissions pass if the same number of rows are returned.
			w.Writeln("return rows.length === records.length;")
			w.Dedent()
			w.Writeln("},")

			w.Dedent()
			w.Writeln("],")
		}
	}

	w.Dedent()
	w.Writeln("}")
}

SELECT
	c.relname::text "table_name",
	a.attname::text "column_name",
	a.attnum "column_num",
	a.attnotnull "not_null",
	a.atthasdef "has_default",
	(
		SELECT pg_catalog.pg_get_expr(d.adbin, d.adrelid, true)
		FROM pg_catalog.pg_attrdef d
		WHERE d.adrelid = a.attrelid AND d.adnum = a.attnum AND a.atthasdef
	) "default_value",
	pg_catalog.format_type(a.atttypid, a.atttypmod) as "data_type"
FROM pg_catalog.pg_attribute a
LEFT JOIN pg_catalog.pg_class c on c.oid = a.attrelid
LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_catalog.pg_index i on i.indexrelid = a.attrelid
WHERE
	n.nspname = 'public'
	AND c.relname not in ('keel_schema', 'keel_refresh_token', 'flow_run', 'flow_step', 'keel_auth_code', 'pg_stat_statements_info', 'pg_stat_statements')
	AND c.relname not like '%__sequence_seq'
	AND a.attnum > 0
	AND NOT a.attisdropped
	AND i.indexrelid is null; -- no indexes

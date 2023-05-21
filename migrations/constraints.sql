SELECT 
	c.relname::text table_name,
	c2.relname::text on_table,
	conname::text constraint_name,
	contype constraint_type,
	confdeltype on_delete ,
	conkey constrained_columns,
	confkey references_columns
FROM pg_catalog.pg_constraint r
LEFT JOIN pg_catalog.pg_class c on c.oid = r.conrelid
LEFT JOIN pg_catalog.pg_class c2 on c2.oid = r.confrelid
LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE 
	n.nspname = 'public'
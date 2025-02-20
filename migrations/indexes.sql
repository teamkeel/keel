SELECT
    i.tablename AS table_name,
    i.indexname AS index_name,
    a.attname AS column_name,
    ix.indisunique AS is_unique,
    ix.indisprimary AS is_primary
FROM pg_catalog.pg_indexes i
JOIN pg_catalog.pg_stat_all_tables t ON i.tablename = t.relname
JOIN pg_catalog.pg_class c ON c.relname = i.indexname
JOIN pg_catalog.pg_index ix ON ix.indexrelid = c.oid
JOIN pg_catalog.pg_attribute a ON a.attrelid = ix.indrelid 
    AND a.attnum = ANY(ix.indkey)
WHERE t.schemaname = 'public'
	AND i.tablename not like ('keel_%')

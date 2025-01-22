SELECT
    routine_name
FROM 
    information_schema.routines
WHERE 
    routine_type = 'FUNCTION'
AND
    routine_schema = 'public' AND routine_name LIKE '%__computed' OR routine_name LIKE '%__computed_dependency';
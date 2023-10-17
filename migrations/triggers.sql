SELECT 
	event_object_table table_name,
	trigger_name trigger_name,
	event_manipulation statement_type,
	action_statement,
	action_timing
FROM 
	information_schema.triggers
WHERE 
	trigger_schema = 'public'
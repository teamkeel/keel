CREATE OR REPLACE FUNCTION process_audit() RETURNS TRIGGER AS $$
DECLARE 
    identity_id_value VARCHAR;
    trace_id_value VARCHAR;
BEGIN
    identity_id_value := nullif(current_setting('audit.identity_id', true), '');
    trace_id_value := nullif(current_setting('audit.trace_id', true ), '');

    IF (TG_OP = 'DELETE') THEN
        INSERT INTO "keel_audit" (table_name, op, data, identity_id, trace_id)
        SELECT TG_TABLE_NAME, 'delete', row_to_json(o.*), identity_id_value, trace_id_value
        FROM old_table o;                                                                 
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO "keel_audit" (table_name, op, data, identity_id, trace_id)                                                                                                                                                                 
        SELECT TG_TABLE_NAME, 'update', row_to_json(n.*), identity_id_value, trace_id_value
        FROM new_table n;                                                                 
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO "keel_audit" (table_name, op, data, identity_id, trace_id)                                                                                                                                                                 
        SELECT TG_TABLE_NAME, 'insert', row_to_json(n.*), identity_id_value, trace_id_value
        FROM new_table n;                                     
    END IF;                                                                                                                                                                              
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
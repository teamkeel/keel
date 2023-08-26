CREATE OR REPLACE FUNCTION process_audit() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO "keel_audit" (table_name, op, data, identity_id, trace_id)
        SELECT TG_TABLE_NAME, 'delete', row_to_json(o.*), current_setting ('audit.identity_id', true ) , current_setting ('audit.trace_id', true )
        FROM old_table o;                                                                 
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO "keel_audit" (table_name, op, data, identity_id, trace_id)                                                                                                                                                                 
        SELECT TG_TABLE_NAME, 'update', row_to_json(n.*), current_setting ('audit.identity_id', true ) , current_setting ('audit.trace_id', true )
        FROM new_table n;                                                                 
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO "keel_audit" (table_name, op, data, identity_id, trace_id)                                                                                                                                                                 
        SELECT TG_TABLE_NAME, 'insert', row_to_json(n.*), current_setting ('audit.identity_id', true ) , current_setting ('audit.trace_id', true )
        FROM new_table n;                                                                 
    END IF;                                                                                                                                                                              
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
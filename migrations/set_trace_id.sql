CREATE OR REPLACE PROCEDURE set_trace_id(id VARCHAR) AS $$
BEGIN
    PERFORM set_config('audit.trace_id', id, true);
END
$$ LANGUAGE plpgsql;
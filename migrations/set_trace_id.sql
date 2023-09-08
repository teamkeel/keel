CREATE OR REPLACE FUNCTION set_trace_id(id VARCHAR) 
RETURNS TEXT AS $$
BEGIN
    RETURN set_config('audit.trace_id', id, true);
END
$$ LANGUAGE plpgsql;
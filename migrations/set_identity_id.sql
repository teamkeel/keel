CREATE OR REPLACE FUNCTION set_identity_id(id VARCHAR)
RETURNS TEXT AS $$
BEGIN
    RETURN set_config('audit.identity_id', id, true);
END
$$ LANGUAGE plpgsql;
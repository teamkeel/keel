CREATE OR REPLACE PROCEDURE set_identity_id(id VARCHAR) AS $$
BEGIN
    PERFORM set_config('audit.identity_id', id, true);
END
$$ LANGUAGE plpgsql;
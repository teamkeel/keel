CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW(); 
    RETURN NEW;
END
$$ LANGUAGE plpgsql;
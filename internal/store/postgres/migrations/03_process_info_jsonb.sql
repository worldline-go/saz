ALTER TABLE ${table_prefix}process
    ALTER COLUMN info TYPE jsonb USING info::jsonb;

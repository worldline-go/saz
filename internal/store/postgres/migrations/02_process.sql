CREATE TABLE IF NOT EXISTS ${table_prefix}process (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    info json NOT NULL,
    "user" TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_${table_prefix}process_status ON ${table_prefix}process (status);

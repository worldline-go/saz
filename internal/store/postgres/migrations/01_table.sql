CREATE TABLE IF NOT EXISTS ${table_prefix}notes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    content json NOT NULL,
    path TEXT NOT NULL UNIQUE,
    updated_by TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_${table_prefix}notes_path ON ${table_prefix}notes(path);

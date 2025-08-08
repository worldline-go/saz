CREATE TABLE IF NOT EXISTS notes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    content json NOT NULL,
    path TEXT NOT NULL UNIQUE,
    updated_by TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notes_path ON notes(path);

CREATE TABLE files (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    folder_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    full_path TEXT NOT NULL,  -- e.g. "Root/Projects/Design/file.pdf"
    s3_key TEXT NOT NULL,     -- Same as full_path or a variation
    size BIGINT,
    mime_type TEXT,
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

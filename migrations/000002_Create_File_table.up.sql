CREATE TABLE files (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    folder_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    full_path TEXT NOT NULL,  -- e.g. "Root/Projects/Design/file.pdf"
    upload_url TEXT,
    s3_key TEXT,     -- Same as full_path or a variation
    size BIGINT,
    mime_type TEXT,
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

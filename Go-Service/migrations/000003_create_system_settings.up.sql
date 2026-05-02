CREATE TABLE IF NOT EXISTS system_settings (
    id                     BIGSERIAL PRIMARY KEY,
    editor_role_id         TEXT     NOT NULL DEFAULT '',
    stream_access_role_ids TEXT[]   NOT NULL DEFAULT '{}'
);

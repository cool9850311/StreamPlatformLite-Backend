CREATE TABLE IF NOT EXISTS livestreams (
    uuid          TEXT     PRIMARY KEY,
    name          TEXT     NOT NULL,
    api_key       TEXT     NOT NULL,
    owner_user_id TEXT     NOT NULL,
    visibility    TEXT     NOT NULL CHECK (visibility IN ('public','member_only','private','link')),
    title         TEXT     NOT NULL DEFAULT '',
    information   TEXT     NOT NULL DEFAULT '',
    ban_list      TEXT[]   NOT NULL DEFAULT '{}',
    mute_list     TEXT[]   NOT NULL DEFAULT '{}',
    is_record     BOOLEAN  NOT NULL DEFAULT false
);
CREATE INDEX IF NOT EXISTS idx_livestreams_owner ON livestreams(owner_user_id);

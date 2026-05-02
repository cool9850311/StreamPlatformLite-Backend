CREATE TABLE IF NOT EXISTS accounts (
    id       BIGSERIAL PRIMARY KEY,
    username TEXT      UNIQUE NOT NULL,
    password TEXT      NOT NULL,
    role     SMALLINT  NOT NULL CHECK (role >= 0 AND role <= 5)
);

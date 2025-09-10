CREATE TABLE IF NOT EXISTS context_entries
(
    id      BIGSERIAL PRIMARY KEY,
    created TIMESTAMP          NOT NULL,
    tags    VARCHAR(255) ARRAY NOT NULL,
    text    TEXT               NOT NULL
);

CREATE TABLE IF NOT EXISTS scheduled_jobs
(
    id   BIGSERIAL PRIMARY KEY,
    data JSON NOT NULL
);

CREATE TABLE IF NOT EXISTS migration
(
    id      VARCHAR(255) PRIMARY KEY,
    applied TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS scheduled_jobs
(
    name    VARCHAR(255) PRIMARY KEY,
    created TIMESTAMP NOT NULL,
    data    JSON      NOT NULL
);

CREATE TABLE IF NOT EXISTS migration
(
    id      VARCHAR(255) PRIMARY KEY,
    applied TIMESTAMP NOT NULL
);

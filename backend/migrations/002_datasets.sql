CREATE TABLE IF NOT EXISTS datasets (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_queries (
    id         BIGSERIAL PRIMARY KEY,
    dataset_id BIGINT      NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    query      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

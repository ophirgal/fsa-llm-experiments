CREATE TYPE experiment_status AS ENUM ('ready', 'in progress', 'done', 'failed');

CREATE TABLE IF NOT EXISTS experiments (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT                NOT NULL,
    dataset_id  BIGINT              NOT NULL REFERENCES datasets(id) ON DELETE RESTRICT,
    status      experiment_status   NOT NULL DEFAULT 'ready',
    total_score DOUBLE PRECISION,
    start_time  TIMESTAMPTZ,
    end_time    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS experiment_prompts (
    id            BIGSERIAL PRIMARY KEY,
    experiment_id BIGINT      NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    prompt        TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS experiment_judge_prompt (
    id            BIGSERIAL PRIMARY KEY,
    experiment_id BIGINT      NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    prompt        TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

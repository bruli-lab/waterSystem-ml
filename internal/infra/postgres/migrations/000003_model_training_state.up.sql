CREATE TABLE IF NOT EXISTS model_training_state
(
    zone             VARCHAR PRIMARY KEY,
    last_training_at TIMESTAMPTZ NOT NULL
);
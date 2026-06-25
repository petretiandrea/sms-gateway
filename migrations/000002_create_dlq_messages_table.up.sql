CREATE TABLE IF NOT EXISTS dlq_messages (
    id BIGSERIAL PRIMARY KEY,
    message_id TEXT NOT NULL,
    channel TEXT NOT NULL,
    exchange TEXT NOT NULL DEFAULT '',
    routing_key TEXT NOT NULL DEFAULT '',
    content_type TEXT NOT NULL DEFAULT '',
    payload BYTEA NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_dlq_messages_created_at
    ON dlq_messages (created_at);

CREATE INDEX IF NOT EXISTS idx_dlq_messages_channel
    ON dlq_messages (channel);

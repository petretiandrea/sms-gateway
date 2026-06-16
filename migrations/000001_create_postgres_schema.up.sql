CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY,
    phone TEXT NOT NULL,
    api_key TEXT NOT NULL UNIQUE,
    is_suspended BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS phones (
    id UUID PRIMARY KEY,
    phone TEXT NOT NULL UNIQUE,
    account_id UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    fcm_token TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_phones_account_id
    ON phones (account_id);

CREATE TABLE IF NOT EXISTS sms_messages (
    id UUID PRIMARY KEY,
    from_phone TEXT NOT NULL,
    to_phone TEXT NOT NULL,
    content TEXT NOT NULL,
    is_sent BOOLEAN NOT NULL DEFAULT FALSE,
    last_attempt_type TEXT NULL,
    last_attempt_phone_id UUID NULL REFERENCES phones (id) ON DELETE SET NULL,
    last_attempt_count INTEGER NULL,
    last_attempt_failure_reason TEXT NULL,
    owner_id UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    webhook_url TEXT NULL,
    CONSTRAINT chk_sms_messages_last_attempt_type
        CHECK (last_attempt_type IS NULL OR last_attempt_type IN ('success', 'failure')),
    CONSTRAINT chk_sms_messages_last_attempt_count
        CHECK (
            (last_attempt_type IS NULL AND last_attempt_count IS NULL)
            OR (last_attempt_type IS NOT NULL AND last_attempt_count IS NOT NULL)
        ),
    CONSTRAINT chk_sms_messages_failure_reason
        CHECK (
            last_attempt_type IS DISTINCT FROM 'failure'
            OR last_attempt_failure_reason IS NOT NULL
        )
);

CREATE INDEX IF NOT EXISTS idx_sms_messages_owner_id
    ON sms_messages (owner_id);

CREATE INDEX IF NOT EXISTS idx_sms_messages_from_phone
    ON sms_messages (from_phone);

CREATE INDEX IF NOT EXISTS idx_sms_messages_is_sent
    ON sms_messages (is_sent);

CREATE INDEX IF NOT EXISTS idx_sms_messages_created_at
    ON sms_messages (created_at);

CREATE TABLE IF NOT EXISTS delivery_notification_configs (
    account_id UUID PRIMARY KEY REFERENCES accounts (id) ON DELETE CASCADE,
    webhook_url TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE
);

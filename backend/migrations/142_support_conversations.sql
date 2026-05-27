CREATE TABLE IF NOT EXISTS support_conversations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    subject VARCHAR(200) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_user_message_at TIMESTAMPTZ NULL,
    last_admin_message_at TIMESTAMPTZ NULL,
    CONSTRAINT support_conversations_status_check CHECK (status IN ('open', 'closed')),
    CONSTRAINT support_conversations_user_order_unique UNIQUE (user_id, order_id)
);

CREATE INDEX IF NOT EXISTS idx_support_conversations_user_last_message
    ON support_conversations (user_id, last_message_at DESC);

CREATE INDEX IF NOT EXISTS idx_support_conversations_status_last_message
    ON support_conversations (status, last_message_at DESC);

CREATE INDEX IF NOT EXISTS idx_support_conversations_order_id
    ON support_conversations (order_id);

CREATE TABLE IF NOT EXISTS support_messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES support_conversations(id) ON DELETE CASCADE,
    sender_type VARCHAR(20) NOT NULL,
    sender_user_id BIGINT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT support_messages_sender_type_check CHECK (sender_type IN ('user', 'admin'))
);

CREATE INDEX IF NOT EXISTS idx_support_messages_conversation_created
    ON support_messages (conversation_id, created_at ASC);

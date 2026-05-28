ALTER TABLE support_conversations
    ADD COLUMN IF NOT EXISTS last_user_read_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS last_admin_read_at TIMESTAMPTZ NULL;

UPDATE support_conversations
SET last_user_read_at = COALESCE(last_user_read_at, last_user_message_at, created_at),
    last_admin_read_at = COALESCE(last_admin_read_at, last_admin_message_at, created_at);

DROP TABLE IF EXISTS delivery_notification_configs;

DROP INDEX IF EXISTS idx_sms_messages_created_at;
DROP INDEX IF EXISTS idx_sms_messages_is_sent;
DROP INDEX IF EXISTS idx_sms_messages_from_phone;
DROP INDEX IF EXISTS idx_sms_messages_owner_id;
DROP TABLE IF EXISTS sms_messages;

DROP INDEX IF EXISTS idx_phones_account_id;
DROP TABLE IF EXISTS phones;

DROP TABLE IF EXISTS accounts;

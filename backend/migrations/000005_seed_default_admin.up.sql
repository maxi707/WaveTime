DELETE FROM users WHERE phone = 'lpgin';

INSERT INTO users (role, full_name, phone, email, password_hash, is_active)
VALUES (
  'admin',
  'Default Admin',
  'admin',
  NULL,
  'sha256$120000$d2F2ZXRpbWVfYWRtaW5fXw$FipHfBs33RBBS9WZKw3INJ2ZeCDN67bu31WVQ4VMqHE',
  TRUE
)
ON CONFLICT (phone)
DO UPDATE SET
  role = 'admin',
  full_name = EXCLUDED.full_name,
  password_hash = EXCLUDED.password_hash,
  is_active = TRUE,
  updated_at = NOW();

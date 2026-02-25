CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  role user_role NOT NULL DEFAULT 'client',
  full_name TEXT NOT NULL,
  phone TEXT UNIQUE,
  email CITEXT UNIQUE,
  password_hash TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (phone IS NOT NULL OR email IS NOT NULL)
);

CREATE TABLE pools (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  address TEXT NOT NULL,
  timezone TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE training_types (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  duration_minutes INTEGER NOT NULL DEFAULT 60,
  default_capacity INTEGER NOT NULL CHECK (default_capacity > 0),
  price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (duration_minutes = 60)
);

CREATE TABLE slots (
  id BIGSERIAL PRIMARY KEY,
  pool_id BIGINT NOT NULL REFERENCES pools (id),
  training_type_id BIGINT NOT NULL REFERENCES training_types (id),
  starts_at TIMESTAMPTZ NOT NULL,
  ends_at TIMESTAMPTZ NOT NULL,
  capacity INTEGER NOT NULL CHECK (capacity > 0),
  booked_count INTEGER NOT NULL DEFAULT 0,
  price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
  status slot_status NOT NULL DEFAULT 'open',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (ends_at = starts_at + INTERVAL '60 minutes'),
  CHECK (booked_count >= 0 AND booked_count <= capacity),
  UNIQUE (pool_id, training_type_id, starts_at)
);

CREATE TABLE bookings (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users (id),
  slot_id BIGINT NOT NULL REFERENCES slots (id),
  status booking_status NOT NULL DEFAULT 'pending_payment',
  reserved_until TIMESTAMPTZ,
  price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
  source TEXT NOT NULL DEFAULT 'web',
  cancel_reason TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, slot_id),
  CHECK (
    (status IN ('pending_payment', 'reserved') AND reserved_until IS NOT NULL)
    OR status NOT IN ('pending_payment', 'reserved')
  )
);

CREATE TABLE payments (
  id BIGSERIAL PRIMARY KEY,
  booking_id BIGINT NOT NULL REFERENCES bookings (id),
  provider TEXT NOT NULL,
  amount NUMERIC(10, 2) NOT NULL CHECK (amount >= 0),
  currency CHAR(3) NOT NULL DEFAULT 'RUB',
  status payment_status NOT NULL DEFAULT 'pending',
  external_id TEXT,
  paid_at TIMESTAMPTZ,
  raw_payload JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (external_id)
);

CREATE TABLE attendance (
  id BIGSERIAL PRIMARY KEY,
  booking_id BIGINT NOT NULL UNIQUE REFERENCES bookings (id),
  checkin_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  marked_by BIGINT REFERENCES users (id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE admin_actions (
  id BIGSERIAL PRIMARY KEY,
  admin_id BIGINT NOT NULL REFERENCES users (id),
  entity TEXT NOT NULL,
  entity_id BIGINT NOT NULL,
  action TEXT NOT NULL,
  reason TEXT,
  before_json JSONB,
  after_json JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notification_queue (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users (id),
  channel notification_channel NOT NULL,
  template TEXT NOT NULL,
  payload JSONB NOT NULL,
  status notification_status NOT NULL DEFAULT 'pending',
  attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  last_error TEXT,
  scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  sent_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE user_role AS ENUM ('client', 'admin');
CREATE TYPE slot_status AS ENUM ('open', 'closed', 'cancelled');
CREATE TYPE booking_status AS ENUM (
  'pending_payment',
  'reserved',
  'confirmed',
  'cancelled',
  'expired',
  'attended',
  'no_show'
);
CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed', 'refunded', 'cancelled');
CREATE TYPE notification_channel AS ENUM ('email', 'sms');
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed', 'cancelled');

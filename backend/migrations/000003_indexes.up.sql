CREATE INDEX idx_slots_pool_starts_at ON slots (pool_id, starts_at);
CREATE INDEX idx_slots_status_starts_at ON slots (status, starts_at);

CREATE INDEX idx_bookings_user_status_created ON bookings (user_id, status, created_at DESC);
CREATE INDEX idx_bookings_slot_status ON bookings (slot_id, status);
CREATE INDEX idx_bookings_reserved_until ON bookings (reserved_until)
  WHERE status IN ('pending_payment', 'reserved');

CREATE INDEX idx_payments_booking_status ON payments (booking_id, status);
CREATE INDEX idx_payments_created_at ON payments (created_at DESC);

CREATE INDEX idx_admin_actions_entity_created ON admin_actions (entity, entity_id, created_at DESC);

CREATE INDEX idx_notification_queue_status_scheduled ON notification_queue (status, scheduled_at);
CREATE INDEX idx_notification_queue_user_created ON notification_queue (user_id, created_at DESC);

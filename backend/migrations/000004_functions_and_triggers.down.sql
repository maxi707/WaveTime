DROP TRIGGER IF EXISTS trg_bookings_sync_slot_booked_count ON bookings;
DROP TRIGGER IF EXISTS trg_notification_queue_set_updated_at ON notification_queue;
DROP TRIGGER IF EXISTS trg_payments_set_updated_at ON payments;
DROP TRIGGER IF EXISTS trg_bookings_set_updated_at ON bookings;
DROP TRIGGER IF EXISTS trg_slots_set_updated_at ON slots;
DROP TRIGGER IF EXISTS trg_training_types_set_updated_at ON training_types;
DROP TRIGGER IF EXISTS trg_pools_set_updated_at ON pools;
DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;

DROP FUNCTION IF EXISTS sync_slot_booked_count();
DROP FUNCTION IF EXISTS decrement_slot_booked_count(BIGINT);
DROP FUNCTION IF EXISTS increment_slot_booked_count(BIGINT);
DROP FUNCTION IF EXISTS is_booking_counted(booking_status);
DROP FUNCTION IF EXISTS set_updated_at();

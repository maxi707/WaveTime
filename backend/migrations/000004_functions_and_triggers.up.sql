CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION is_booking_counted(input_status booking_status)
RETURNS BOOLEAN
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT input_status IN ('pending_payment', 'reserved', 'confirmed', 'attended', 'no_show');
$$;

CREATE OR REPLACE FUNCTION increment_slot_booked_count(target_slot_id BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
  UPDATE slots
  SET booked_count = booked_count + 1
  WHERE id = target_slot_id
    AND status = 'open'
    AND booked_count < capacity;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Cannot reserve slot %, it is full or unavailable', target_slot_id;
  END IF;
END;
$$;

CREATE OR REPLACE FUNCTION decrement_slot_booked_count(target_slot_id BIGINT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
  UPDATE slots
  SET booked_count = booked_count - 1
  WHERE id = target_slot_id
    AND booked_count > 0;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Cannot decrement slot % booked_count', target_slot_id;
  END IF;
END;
$$;

CREATE OR REPLACE FUNCTION sync_slot_booked_count()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  old_is_counted BOOLEAN;
  new_is_counted BOOLEAN;
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF is_booking_counted(NEW.status) THEN
      PERFORM increment_slot_booked_count(NEW.slot_id);
    END IF;
    RETURN NEW;
  END IF;

  IF TG_OP = 'UPDATE' THEN
    old_is_counted := is_booking_counted(OLD.status);
    new_is_counted := is_booking_counted(NEW.status);

    IF OLD.slot_id = NEW.slot_id THEN
      IF old_is_counted AND NOT new_is_counted THEN
        PERFORM decrement_slot_booked_count(OLD.slot_id);
      ELSIF NOT old_is_counted AND new_is_counted THEN
        PERFORM increment_slot_booked_count(NEW.slot_id);
      END IF;
    ELSE
      IF old_is_counted THEN
        PERFORM decrement_slot_booked_count(OLD.slot_id);
      END IF;
      IF new_is_counted THEN
        PERFORM increment_slot_booked_count(NEW.slot_id);
      END IF;
    END IF;

    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    IF is_booking_counted(OLD.status) THEN
      PERFORM decrement_slot_booked_count(OLD.slot_id);
    END IF;
    RETURN OLD;
  END IF;

  RETURN NULL;
END;
$$;

CREATE TRIGGER trg_users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_pools_set_updated_at
BEFORE UPDATE ON pools
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_training_types_set_updated_at
BEFORE UPDATE ON training_types
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_slots_set_updated_at
BEFORE UPDATE ON slots
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_bookings_set_updated_at
BEFORE UPDATE ON bookings
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_payments_set_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_notification_queue_set_updated_at
BEFORE UPDATE ON notification_queue
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_bookings_sync_slot_booked_count
BEFORE INSERT OR UPDATE OR DELETE ON bookings
FOR EACH ROW
EXECUTE FUNCTION sync_slot_booked_count();

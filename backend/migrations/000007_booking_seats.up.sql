ALTER TABLE bookings
ADD COLUMN IF NOT EXISTS seats_count INTEGER NOT NULL DEFAULT 1 CHECK (seats_count > 0);

CREATE OR REPLACE FUNCTION increment_slot_booked_count(target_slot_id BIGINT, seats INTEGER)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
  IF seats <= 0 THEN
    RAISE EXCEPTION 'Seats increment must be positive, got %', seats;
  END IF;

  UPDATE slots
  SET booked_count = booked_count + seats
  WHERE id = target_slot_id
    AND status = 'open'
    AND booked_count + seats <= capacity;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Cannot reserve % seat(s) for slot %, it is full or unavailable', seats, target_slot_id;
  END IF;
END;
$$;

CREATE OR REPLACE FUNCTION decrement_slot_booked_count(target_slot_id BIGINT, seats INTEGER)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
  IF seats <= 0 THEN
    RAISE EXCEPTION 'Seats decrement must be positive, got %', seats;
  END IF;

  UPDATE slots
  SET booked_count = booked_count - seats
  WHERE id = target_slot_id
    AND booked_count - seats >= 0;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Cannot decrement % seat(s) for slot %', seats, target_slot_id;
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
  old_seats INTEGER;
  new_seats INTEGER;
  delta INTEGER;
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF is_booking_counted(NEW.status) THEN
      PERFORM increment_slot_booked_count(NEW.slot_id, NEW.seats_count);
    END IF;
    RETURN NEW;
  END IF;

  IF TG_OP = 'UPDATE' THEN
    old_is_counted := is_booking_counted(OLD.status);
    new_is_counted := is_booking_counted(NEW.status);
    old_seats := COALESCE(OLD.seats_count, 1);
    new_seats := COALESCE(NEW.seats_count, 1);

    IF OLD.slot_id = NEW.slot_id THEN
      IF old_is_counted AND new_is_counted THEN
        delta := new_seats - old_seats;
        IF delta > 0 THEN
          PERFORM increment_slot_booked_count(NEW.slot_id, delta);
        ELSIF delta < 0 THEN
          PERFORM decrement_slot_booked_count(OLD.slot_id, -delta);
        END IF;
      ELSIF old_is_counted AND NOT new_is_counted THEN
        PERFORM decrement_slot_booked_count(OLD.slot_id, old_seats);
      ELSIF NOT old_is_counted AND new_is_counted THEN
        PERFORM increment_slot_booked_count(NEW.slot_id, new_seats);
      END IF;
    ELSE
      IF old_is_counted THEN
        PERFORM decrement_slot_booked_count(OLD.slot_id, old_seats);
      END IF;
      IF new_is_counted THEN
        PERFORM increment_slot_booked_count(NEW.slot_id, new_seats);
      END IF;
    END IF;

    RETURN NEW;
  END IF;

  IF TG_OP = 'DELETE' THEN
    IF is_booking_counted(OLD.status) THEN
      PERFORM decrement_slot_booked_count(OLD.slot_id, COALESCE(OLD.seats_count, 1));
    END IF;
    RETURN OLD;
  END IF;

  RETURN NULL;
END;
$$;

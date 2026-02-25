-- Rollback for test month schedule seed.
-- Deletes generated-like slots for the same moving 1-month window.
-- Keep this for non-production/test environments.

WITH bounds AS (
  SELECT current_date::date AS date_from,
         (current_date + INTERVAL '1 month')::date AS date_to
)
DELETE FROM slots s
USING pools p, training_types t, bounds b
WHERE s.pool_id = p.id
  AND s.training_type_id = t.id
  AND p.is_active = TRUE
  AND t.is_active = TRUE
  AND s.status = 'open'
  AND s.starts_at >= b.date_from::timestamptz
  AND s.starts_at < b.date_to::timestamptz
  AND EXTRACT(MINUTE FROM s.starts_at) = 0
  AND EXTRACT(SECOND FROM s.starts_at) = 0
  AND s.ends_at = s.starts_at + INTERVAL '60 minutes'
  AND s.capacity = t.default_capacity
  AND s.price = t.price;

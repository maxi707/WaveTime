-- Seed test schedule for one month from current_date (inclusive).
-- For each active pool and active training type:
-- daily slots 07:00..22:00 (60 minutes, ending at 23:00).

WITH bounds AS (
  SELECT current_date::date AS date_from,
         (current_date + INTERVAL '1 month')::date AS date_to
),
days AS (
  SELECT gs::date AS d
  FROM bounds b
  CROSS JOIN generate_series(b.date_from, b.date_to - INTERVAL '1 day', INTERVAL '1 day') gs
),
hours AS (
  SELECT generate_series(7, 22) AS h
),
seed AS (
  SELECT
    p.id AS pool_id,
    t.id AS training_type_id,
    ((d.d::timestamp + make_interval(hours => h.h)) AT TIME ZONE p.timezone) AS starts_at,
    (((d.d::timestamp + make_interval(hours => h.h)) AT TIME ZONE p.timezone) + INTERVAL '60 minutes') AS ends_at,
    t.default_capacity AS capacity,
    t.price AS price
  FROM pools p
  CROSS JOIN training_types t
  CROSS JOIN days d
  CROSS JOIN hours h
  WHERE p.is_active = TRUE
    AND t.is_active = TRUE
)
INSERT INTO slots (pool_id, training_type_id, starts_at, ends_at, capacity, price, status)
SELECT pool_id, training_type_id, starts_at, ends_at, capacity, price, 'open'
FROM seed
ON CONFLICT (pool_id, training_type_id, starts_at) DO NOTHING;
